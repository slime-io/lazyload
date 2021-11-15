/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	prometheusApi "github.com/prometheus/client_golang/api"
	prometheusV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"slime.io/slime/framework/apis/config/v1alpha1"
	"slime.io/slime/framework/model/metric"
	"slime.io/slime/framework/model/trigger"
	"sort"
	"strings"
	"sync"
	"time"

	"slime.io/slime/framework/model"

	istio "istio.io/api/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slime.io/slime/framework/apis/networking/v1alpha3"
	"slime.io/slime/framework/bootstrap"
	"slime.io/slime/framework/controllers"
	"slime.io/slime/framework/util"

	stderrors "errors"
	lazyloadv1alpha1 "slime.io/slime/modules/lazyload/api/v1alpha1"
	modmodel "slime.io/slime/modules/lazyload/model"
)

// ServicefenceReconciler reconciles a Servicefence object
type ServicefenceReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	cfg               *v1alpha1.Fence
	env               bootstrap.Environment
	interestMeta      map[string]bool
	interestMetaCopy  map[string]bool // for outside read
	watcherMetricChan <-chan metric.Metric
	tickerMetricChan  <-chan metric.Metric
	reconcileLock     sync.RWMutex
	staleNamespaces   map[string]bool
	enabledNamespaces map[string]bool
}

// NewReconciler returns a new reconcile.Reconciler
func NewReconciler(cfg *v1alpha1.Fence, mgr manager.Manager, env bootstrap.Environment) *ServicefenceReconciler {
	log := modmodel.ModuleLog.WithField(model.LogFieldKeyFunction, "NewReconciler")

	// generate producer config
	pc, err := newProducerConfig(env)
	if err != nil {
		log.Errorf("%v", err)
		return nil
	}

	r := &ServicefenceReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		env:               env,
		interestMeta:      map[string]bool{},
		interestMetaCopy:  map[string]bool{},
		watcherMetricChan: pc.WatcherProducerConfig.MetricChan,
		tickerMetricChan:  pc.TickerProducerConfig.MetricChan,
		staleNamespaces:   map[string]bool{},
		enabledNamespaces: map[string]bool{},
	}

	// reconciler defines producer metric handler
	pc.WatcherProducerConfig.NeedUpdateMetricHandler = r.handleWatcherEvent
	pc.TickerProducerConfig.NeedUpdateMetricHandler = r.handleTickerEvent

	// start producer
	metric.NewProducer(pc)
	log.Infof("producers starts")

	if env.Config.Metric != nil {
		go r.WatchMetric()
	}

	return r
}

// +kubebuilder:rbac:groups=microservice.slime.io,resources=servicefences,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=microservice.slime.io,resources=servicefences/status,verbs=get;update;patch

func (r *ServicefenceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	log := modmodel.ModuleLog.WithField(model.LogFieldKeyResource, req.NamespacedName)

	// Fetch the ServiceFence instance
	instance := &lazyloadv1alpha1.ServiceFence{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)

	r.reconcileLock.Lock()
	defer r.reconcileLock.Unlock()

	// TODO watch sidecar
	if err != nil {
		if errors.IsNotFound(err) {
			// TODO should be recovered? maybe we should call refreshFenceStatusOfService here
			log.Info("serviceFence is deleted")
			//r.interestMeta.Pop(req.NamespacedName.String())
			delete(r.interestMeta, req.NamespacedName.String())
			r.updateInterestMetaCopy()
			return r.refreshFenceStatusOfService(context.TODO(), nil, req.NamespacedName)
		} else {
			log.Errorf("get serviceFence error,%+v", err)
			return reconcile.Result{}, err
		}
	}

	if rev := model.IstioRevFromLabel(instance.Labels); !r.env.RevInScope(rev) { // remove watch ?
		log.Infof("exsiting sf %v istioRev %s but our %s, skip...",
			req.NamespacedName, rev, r.env.IstioRev())
		return reconcile.Result{}, nil
	}
	log.Infof("ServicefenceReconciler got serviceFence request, %+v", req.NamespacedName)

	// 资源更新
	diff := r.updateVisitedHostStatus(instance)
	r.recordVisitor(instance, diff)
	if instance.Spec.Enable {
		r.interestMeta[req.NamespacedName.String()] = true
		r.updateInterestMetaCopy()
		err = r.refreshSidecar(instance)
	}

	return ctrl.Result{}, err
}

func (r *ServicefenceReconciler) updateInterestMetaCopy() {
	newInterestMeta := make(map[string]bool)
	for k, v := range r.interestMeta {
		newInterestMeta[k] = v
	}
	r.interestMetaCopy = newInterestMeta
}

func (r *ServicefenceReconciler) getInterestMeta() map[string]bool {
	r.reconcileLock.RLock()
	defer r.reconcileLock.RUnlock()
	return r.interestMetaCopy
}

// call back function for watcher producer
func (r *ServicefenceReconciler) handleWatcherEvent(event trigger.WatcherEvent) metric.QueryMap {

	// check event
	gvks := []schema.GroupVersionKind{
		{Group: "networking.istio.io", Version: "v1beta1", Kind: "Sidecar"},
	}
	invalidEvent := false
	for _, gvk := range gvks {
		if event.GVK == gvk && r.getInterestMeta()[event.NN.String()] {
			invalidEvent = true
		}
	}
	if !invalidEvent {
		return nil
	}

	// generate query map for producer
	qm := make(map[string][]metric.Handler)
	var hs []metric.Handler
	for pName, pHandler := range r.env.Config.Metric.Prometheus.Handlers {
		hs = append(hs, generateHandler(event.NN.Name, event.NN.Namespace, pName, pHandler))
	}
	qm[event.NN.String()] = hs
	return qm
}

// call back function for ticker producer
func (r *ServicefenceReconciler) handleTickerEvent(event trigger.TickerEvent) metric.QueryMap {

	// no need to check time duration

	// generate query map for producer
	qm := make(map[string][]metric.Handler)
	for meta := range r.getInterestMeta() {
		namespace, name := strings.Split(meta, "/")[0], strings.Split(meta, "/")[1]
		var hs []metric.Handler
		for pName, pHandler := range r.env.Config.Metric.Prometheus.Handlers {
			hs = append(hs, generateHandler(name, namespace, pName, pHandler))
		}
		qm[meta] = hs
	}

	return qm
}

func generateHandler(name, namespace, pName string, pHandler *v1alpha1.Prometheus_Source_Handler) metric.Handler {
	query := strings.ReplaceAll(pHandler.Query, "$namespace", namespace)
	query = strings.ReplaceAll(query, "$source_app", name)
	return metric.Handler{Name: pName, Query: query}
}

func newProducerConfig(env bootstrap.Environment) (*metric.ProducerConfig, error) {

	prometheusSourceConfig, err := newPrometheusSourceConfig(env)
	if err != nil {
		return nil, err
	}

	return &metric.ProducerConfig{
		EnableWatcherProducer: true,
		WatcherProducerConfig: metric.WatcherProducerConfig{
			Name:       "lazyload-watcher",
			MetricChan: make(chan metric.Metric),
			WatcherTriggerConfig: trigger.WatcherTriggerConfig{
				Kinds: []schema.GroupVersionKind{
					{
						Group:   "networking.istio.io",
						Version: "v1beta1",
						Kind:    "Sidecar",
					},
				},
				EventChan:     make(chan trigger.WatcherEvent),
				DynamicClient: env.DynamicClient,
			},
			PrometheusSourceConfig: prometheusSourceConfig,
		},
		EnableTickerProducer: true,
		TickerProducerConfig: metric.TickerProducerConfig{
			Name:       "lazyload-ticker",
			MetricChan: make(chan metric.Metric),
			TickerTriggerConfig: trigger.TickerTriggerConfig{
				Durations: []time.Duration{
					30 * time.Second,
				},
				EventChan: make(chan trigger.TickerEvent),
			},
			PrometheusSourceConfig: prometheusSourceConfig,
		},
		StopChan: env.Stop,
	}, nil

}

func newPrometheusSourceConfig(env bootstrap.Environment) (metric.PrometheusSourceConfig, error) {
	ps := env.Config.Metric.Prometheus
	if ps == nil {
		return metric.PrometheusSourceConfig{}, stderrors.New("failure create prometheus client, empty prometheus config")
	}
	promClient, err := prometheusApi.NewClient(prometheusApi.Config{
		Address:      ps.Address,
		RoundTripper: nil,
	})
	if err != nil {
		return metric.PrometheusSourceConfig{}, err
	}

	return metric.PrometheusSourceConfig{
		Api: prometheusV1.NewAPI(promClient),
	}, nil
}

func (r *ServicefenceReconciler) refreshSidecar(instance *lazyloadv1alpha1.ServiceFence) error {
	log := log.WithField("reporter", "ServicefenceReconciler").WithField("function", "refreshSidecar")
	sidecar, err := newSidecar(instance, r.env)
	if err != nil {
		log.Errorf("servicefence generate sidecar failed, %+v", err)
		return err
	}
	if sidecar == nil {
		return nil
	}
	// Set VisitedHost instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, sidecar, r.Scheme); err != nil {
		log.Errorf("attach ownerReference to sidecar failed, %+v", err)
		return err
	}
	sfRev := model.IstioRevFromLabel(instance.Labels)
	model.PatchIstioRevLabel(&sidecar.Labels, sfRev)

	// Check if this Pod already exists
	found := &v1alpha3.Sidecar{}
	nsName := types.NamespacedName{Name: sidecar.Name, Namespace: sidecar.Namespace}
	err = r.Client.Get(context.TODO(), nsName, found)

	if err != nil {
		if errors.IsNotFound(err) {
			found = nil
			err = nil
		} else {
			return err
		}
	}

	if found == nil {
		log.Infof("Creating a new Sidecar in %s:%s", sidecar.Namespace, sidecar.Name)
		err = r.Client.Create(context.TODO(), sidecar)
		if err != nil {
			return err
		}
	} else if rev := model.IstioRevFromLabel(found.Labels); rev != sfRev {
		log.Infof("existed sidecar %v istioRev %s but our rev %s, skip update ...",
			nsName, rev, sfRev)
	} else {
		if !reflect.DeepEqual(found.Spec, sidecar.Spec) {
			log.Infof("Update a Sidecar in %s:%s", sidecar.Namespace, sidecar.Name)
			sidecar.ResourceVersion = found.ResourceVersion
			err = r.Client.Update(context.TODO(), sidecar)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// recordVisitor update the dest servicefences' visitor according to src sf's visit diff
func (r *ServicefenceReconciler) recordVisitor(sf *lazyloadv1alpha1.ServiceFence, diff Diff) {
	for _, addHost := range diff.Added {
		destSf := r.prepareDestFence(sf, addHost)
		if destSf == nil {
			continue
		}
		destSf.Status.Visitor[sf.Namespace+"/"+sf.Name] = true
		_ = r.Client.Status().Update(context.TODO(), destSf)
	}

	for _, delHost := range diff.Deleted {
		destSf := r.prepareDestFence(sf, delHost)
		if destSf == nil {
			continue
		}
		delete(destSf.Status.Visitor, sf.Namespace+"/"+sf.Name)
		_ = r.Client.Status().Update(context.TODO(), destSf)
	}
}

// prepareDestFence prepares servicefence of specified host
func (r *ServicefenceReconciler) prepareDestFence(srcSf *lazyloadv1alpha1.ServiceFence, h string) *lazyloadv1alpha1.ServiceFence {
	nsName := parseHost(srcSf.Namespace, h)
	if nsName == nil {
		return nil
	}

	svc := &corev1.Service{}
	if err := r.Client.Get(context.TODO(), *nsName, svc); err != nil {
		// XXX err handle
		return nil
	}

	destSf := &lazyloadv1alpha1.ServiceFence{}
retry: // FIXME fix infinite loop
	err := r.Client.Get(context.TODO(), *nsName, destSf)
	if err != nil {
		if errors.IsNotFound(err) {
			// XXX maybe should not auto create
			destSf.Name = nsName.Name
			destSf.Namespace = nsName.Namespace
			// XXX set controlled by
			// XXX patch rev
			if err = r.Client.Create(context.TODO(), destSf); err != nil {
				goto retry
			}
		} else {
			return nil
		}
	}

	if destSf.Status.Visitor == nil {
		destSf.Status.Visitor = make(map[string]bool)
	}
	return destSf
}

func parseHost(sourceNs, h string) *types.NamespacedName {
	s := strings.Split(h, ".")
	if len(s) == 5 || len(s) == 2 { // shortname.ns or full(shortname.ns.svc.cluster.local)
		return &types.NamespacedName{
			Namespace: s[1],
			Name:      s[0],
		}
	}
	if len(s) == 1 { // shortname
		return &types.NamespacedName{
			Namespace: sourceNs,
			Name:      s[0],
		}
	}
	return nil // unknown host format, maybe external host
}

func (r *ServicefenceReconciler) updateVisitedHostStatus(host *lazyloadv1alpha1.ServiceFence) Diff {
	checkStatus := func(now int64, strategy *lazyloadv1alpha1.RecyclingStrategy) lazyloadv1alpha1.Destinations_Status {
		switch {
		case strategy.Stable != nil:
			// ...
		case strategy.Deadline != nil:
			if now > strategy.Deadline.Expire.Seconds {
				return lazyloadv1alpha1.Destinations_EXPIRE
			}
		case strategy.Auto != nil:
			if strategy.RecentlyCalled != nil {
				if now-strategy.RecentlyCalled.Seconds > strategy.Auto.Duration.Seconds {
					return lazyloadv1alpha1.Destinations_EXPIRE
				}
			}
		}
		return lazyloadv1alpha1.Destinations_ACTIVE
	}

	domains := make(map[string]*lazyloadv1alpha1.Destinations)
	now := time.Now().Unix()

	// update status of destinations of hosts with config
	for h, strategy := range host.Spec.Host {
		allHost := []string{h}
		if hs := getDestination(h); len(hs) > 0 {
			allHost = append(allHost, hs...)
		}

		domains[h] = &lazyloadv1alpha1.Destinations{
			Hosts:  allHost,
			Status: checkStatus(now, strategy),
		}
	}

	// update status with metric status
	// -> (visited)host -> destinations
	for metricName := range host.Status.MetricStatus {
		metricName = strings.Trim(metricName, "{}")
		if !strings.HasPrefix(metricName, "destination_service") && !strings.HasPrefix(metricName, "request_host") {
			continue
		}
		// destination_service format like: "grafana.istio-system.svc.cluster.local"

		var h, fullHost string
		// trim ""
		if ss := strings.Split(metricName, "\""); len(ss) != 3 {
			continue
		} else {
			// remove port
			h = strings.SplitN(ss[1], ":", 2)[0]
		}

		fullHost = h
		{
			subDomains := strings.Split(h, ".")
			if len(subDomains) == 1 { // shortname
				fullHost = fmt.Sprintf("%s.%s.svc.cluster.local", subDomains[0], host.Namespace)
			} else if len(subDomains) == 2 { // shortname.ns
				fullHost = fmt.Sprintf("%s.%s.svc.cluster.local", subDomains[0], subDomains[1])
			}
		}

		if !isValidHost(fullHost) {
			continue
		}
		if domains[fullHost] != nil { // XXX merge with status from config
			continue
		}

		allHost := []string{fullHost}
		if hs := getDestination(fullHost); len(hs) > 0 {
			allHost = append(allHost, hs...)
		}

		domains[h] = &lazyloadv1alpha1.Destinations{
			Hosts:  allHost,
			Status: lazyloadv1alpha1.Destinations_ACTIVE,
		}
	}

	delta := Diff{
		Deleted: make([]string, 0),
		Added:   make([]string, 0),
	}
	for k := range host.Status.Domains {
		if _, ok := domains[k]; !ok {
			delta.Deleted = append(delta.Deleted, k)
		}
	}
	for k := range domains {
		if _, ok := host.Status.Domains[k]; !ok {
			delta.Added = append(delta.Added, k)
		}
	}
	host.Status.Domains = domains

	_ = r.Client.Status().Update(context.TODO(), host)

	return delta
}

func newSidecar(vhost *lazyloadv1alpha1.ServiceFence, env bootstrap.Environment) (*v1alpha3.Sidecar, error) {
	host := make([]string, 0)
	if !vhost.Spec.Enable {
		return nil, nil
	}
	for _, v := range vhost.Status.Domains {
		if v.Status == lazyloadv1alpha1.Destinations_ACTIVE {
			for _, h := range v.Hosts {
				host = append(host, "*/"+h)
			}
		}
	}
	// sort host to avoid map range random feature resulting in sidecar constant updates
	sort.Strings(host)

	// 需要加入一条根namespace的策略
	host = append(host, env.Config.Global.IstioNamespace+"/*")
	host = append(host, env.Config.Global.SlimeNamespace+"/*")
	// check whether using namespace global-sidecar
	// if so, init config of sidecar will adds */global-sidecar.${svf.ns}.svc.cluster.local
	if env.Config.Global.Misc["global-sidecar-mode"] == "namespace" {
		host = append(host, fmt.Sprintf("*/global-sidecar.%s.svc.cluster.local", vhost.Namespace))
	}
	sidecar := &istio.Sidecar{
		WorkloadSelector: &istio.WorkloadSelector{
			Labels: map[string]string{env.Config.Global.Service: vhost.Name},
		},
		Egress: []*istio.IstioEgressListener{
			{
				// Bind:  "0.0.0.0",
				Hosts: host,
			},
		},
	}

	spec, err := util.ProtoToMap(sidecar)
	if err != nil {
		return nil, err
	}
	ret := &v1alpha3.Sidecar{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vhost.Name,
			Namespace: vhost.Namespace,
		},
		Spec: spec,
	}
	return ret, nil
}

func (r *ServicefenceReconciler) Subscribe(host string, destination interface{}) { // FIXME not used?
	svc, namespace, ok := util.IsK8SService(host)
	if !ok {
		return
	}

	sf := &lazyloadv1alpha1.ServiceFence{}
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: svc, Namespace: namespace}, sf); err != nil {
		return
	}
	istioRev := model.IstioRevFromLabel(sf.Labels)
	if !r.env.RevInScope(istioRev) {
		return
	}

	for k := range sf.Status.Visitor {
		i := strings.Index(k, "/")
		if i < 0 {
			continue
		}

		visitorSf := &lazyloadv1alpha1.ServiceFence{}
		visitorNamespace := k[:i]
		visitorName := k[i+1:]

		err := r.Client.Get(context.TODO(), types.NamespacedName{Name: visitorName, Namespace: visitorNamespace}, visitorSf)
		if err != nil {
			return
		}
		visitorRev := model.IstioRevFromLabel(visitorSf.Labels)
		if !r.env.RevInScope(visitorRev) {
			return
		}

		r.updateVisitedHostStatus(visitorSf)

		sidecar, err := newSidecar(visitorSf, r.env)
		if sidecar == nil {
			continue
		}
		if err != nil {
			return
		}

		// Set VisitedHost instance as the owner and controller
		if err := controllerutil.SetControllerReference(visitorSf, sidecar, r.Scheme); err != nil {
			return
		}
		model.PatchIstioRevLabel(&sidecar.Labels, visitorRev)

		// Check for existence
		found := &v1alpha3.Sidecar{}
		nsName := types.NamespacedName{Name: sidecar.Name, Namespace: sidecar.Namespace}
		err = r.Client.Get(context.TODO(), nsName, found)
		if err != nil {
			if errors.IsNotFound(err) {
				found = nil
				err = nil
			} else {
				return
			}
		}

		if found == nil {
			err = r.Client.Create(context.TODO(), sidecar)
			if err != nil {
				log.Errorf("create sidecar %+v met err %v", sidecar, err)
			}
		} else if scRev := model.IstioRevFromLabel(found.Labels); scRev != visitorRev {
			log.Infof("existing sidecar %v istioRev %s but our %s, skip ...",
				nsName, scRev, visitorRev)
		} else {
			if !reflect.DeepEqual(found.Spec, sidecar.Spec) {
				sidecar.ResourceVersion = found.ResourceVersion
				err = r.Client.Update(context.TODO(), sidecar)

				if err != nil {
					log.Errorf("create sidecar %+v met err %v", sidecar, err)
				}
			}
		}
		return
	}
	return
}

func getDestination(k string) []string {
	if i := controllers.HostDestinationMapping.Get(k); i != nil {
		if hs, ok := i.([]string); ok {
			return hs
		}
	}
	return nil
}

// TODO: More rigorous verification
func isValidHost(h string) bool {
	if strings.Contains(h, "global-sidecar") ||
		strings.Contains(h, ":") ||
		strings.Contains(h, "unknown") {
		return false
	}
	return true
}

func (r *ServicefenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lazyloadv1alpha1.ServiceFence{}).
		Complete(r)
}
