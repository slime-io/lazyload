package module

import (
	"os"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	istioapi "slime.io/slime/framework/apis"
	"slime.io/slime/framework/apis/config/v1alpha1"
	"slime.io/slime/framework/bootstrap"
	basecontroller "slime.io/slime/framework/controllers"
	"slime.io/slime/framework/model"
	lazyloadapiv1alpha1 "slime.io/slime/modules/lazyload/api/v1alpha1"
	"slime.io/slime/modules/lazyload/controllers"
)

const Name = "lazyload"

type Module struct {
	config v1alpha1.Fence
}

func (m *Module) Config() proto.Message {
	return &m.config
}

func (m *Module) Name() string {
	return Name
}

func (m *Module) InitScheme(scheme *runtime.Scheme) error {
	for _, f := range []func(*runtime.Scheme) error{
		clientgoscheme.AddToScheme,
		lazyloadapiv1alpha1.AddToScheme,
		istioapi.AddToScheme,
	} {
		if err := f(scheme); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) InitManager(mgr manager.Manager, env bootstrap.Environment, cbs model.ModuleInitCallbacks) error {
	cfg := &m.config
	if env.Config != nil && env.Config.Fence != nil {
		cfg = env.Config.Fence
	}
	sfReconciler := controllers.NewReconciler(cfg, mgr, &env)

	var builder basecontroller.ObjectReconcilerBuilder
	if err := builder.Add(basecontroller.ObjectReconcileItem{
		Name: "ServiceFence",
		R:    sfReconciler,
	}).Add(basecontroller.ObjectReconcileItem{
		Name: "VirtualService",
		R: &basecontroller.VirtualServiceReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		},
	}).Add(basecontroller.ObjectReconcileItem{
		Name:    "Service",
		ApiType: &corev1.Service{},
		R:       reconcile.Func(sfReconciler.ReconcileService),
	}).Add(basecontroller.ObjectReconcileItem{
		Name:    "Namespace",
		ApiType: &corev1.Namespace{},
		R:       reconcile.Func(sfReconciler.ReconcileNamespace),
	}).Build(mgr); err != nil {
		log.Errorf("unable to create controller,%+v", err)
		os.Exit(1)
	}

	return nil
}
