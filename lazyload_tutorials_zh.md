- [懒加载教程](#懒加载教程)
  - [架构](#架构)
  - [安装和使用](#安装和使用)
    - [使用namespace级别的global-sidecar](#使用namespace级别的global-sidecar)
    - [使用集群唯一的global-sidecar](#使用集群唯一的global-sidecar)
    - [不使用global-sidecar组件](#不使用global-sidecar组件)
  - [其他特性介绍](#其他特性介绍)
    - [基于Accesslog开启懒加载](#基于accesslog开启懒加载)
    - [基于namespace/service label自动生成ServiceFence](#基于namespaceservice-label自动生成servicefence)
    - [自定义兜底流量分派](#自定义兜底流量分派)
    - [日志输出到本地并轮转](#日志输出到本地并轮转)
      - [创建存储卷](#创建存储卷)
      - [在SlimeBoot中声明挂载信息](#在slimeboot中声明挂载信息)
  - [示例](#示例)
    - [安装 istio (1.8+)](#安装-istio-18)
    - [设定tag](#设定tag)
    - [安装 slime](#安装-slime)
    - [安装bookinfo](#安装bookinfo)
    - [开启懒加载](#开启懒加载)
    - [首次访问观察](#首次访问观察)
    - [再次访问观察](#再次访问观察)
    - [卸载](#卸载)
    - [补充说明](#补充说明)
  - [FAQ](#faq)
    - [global-sidecar-pilot的意义？](#global-sidecar-pilot的意义)
    - [global sidecar不能正常启动？](#global-sidecar不能正常启动)


# 懒加载教程

## 架构



![](./media/lazyload-architecture-2021-10-19.png)

注：绿色为lazyload controller内部逻辑

过程说明：

0. 部署lazyload，会创建出lazyload, global-sidecar-pilot, global-sidecar等应用，给特定服务启用懒加载时，还会创建servicefence以及sidecar

1. global-sidecar-pilot从集群原有pilot处获取全量配置信息

2. global-sidecar-pilot针对global-sidecar的特殊需求，将通用配置调整成global-sidecar定制版后，推送配置给global-sidecar

3. 为应用A启用懒加载

   3.1 创建servicefence A

   3.2 创建sidecarsidecar A

   3.3 ApiServer 感知到sidecar A创建

4. Pilot从ApiServer获取Sidecar A的内容

5. Pilot下发sidecar限定范围内的配置给App A的envoy

6. App A发起访问App B，由于envoy没有App B的配置信息，请求经兜底路由，转发到global-sidecar

7. global-sidecar拥有全量配置，正确转发请求到App B

8. global-sidecar通过access log或者prometheus metric等方式上报调用关系App A->App B 

9. lazyload controller获取到调用关系

10. lazyload更新懒加载配置

    10.1 更新servicefence A，添加关于B的metric

    10.2 更新sidecarsidecar A，egress.hosts添加B信息

    10.3 ApiServer 感知到sidecar A更新

11. Pilot从ApiServer获取Sidecar A新内容

12. Pilot下发sidecar限定范围内的配置给App A的envoy，新增了B的xDS内容

13. A直接访问B成功



## 安装和使用

### 使用namespace级别的global-sidecar

请先按照安装slime-boot小节的指引安装slime-boot

1. 使用Slime的配置懒加载功能需打开Fence模块，同时安装global-sidecar, pilot等附加组件，如下：

   > [完整样例](./install/samples/lazyload/slimeboot_lazyload.yaml)

```yaml
apiVersion: config.netease.com/v1alpha1
kind: SlimeBoot
metadata:
  name: lazyload
  namespace: mesh-operator
spec:
  image:
    pullPolicy: Always
    repository: docker.io/slimeio/slime-lazyload
    tag: {{your_lazyload_tag}}
  module:
    - name: lazyload
      fence:
        enable: true
        wormholePort: 
          - "{{your_port}}" # replace to your application service ports, and extend the list in case of multi ports
      metric:
        prometheus:
          address: {{prometheus_address}} # replace to your prometheus address
          handlers:
            destination:
              query: |
                sum(istio_requests_total{source_app="$source_app",reporter="destination"})by(destination_service)
              type: Group
  component:
    globalSidecar:
      enable: true
      type: namespaced
      namespace:
        - {{your_namespace}} # replace to your service's namespace, and extend the list in case of multi namespaces
      resources:
        requests:
          cpu: 200m
          memory: 200Mi
        limits:
          cpu: 200m
          memory: 200Mi
      image:
        repository: {{your_sidecar_repo}}
        tag: {{your_sidecar_tag}}          
    pilot:
      enable: true
      resources:
        requests:
          cpu: 200m
          memory: 200Mi
        limits:
          cpu: 200m
          memory: 200Mi
      image:
        repository: {{your_pilot_repo}}
        tag: {{your_pilot_tag}}
```



​	2.确认所有组件已正常运行：

```
$ kubectl get po -n mesh-operator
NAME                                    READY     STATUS    RESTARTS   AGE
global-sidecar-pilot-796fb554d7-blbml   1/1       Running   0          27s
lazyload-fbcd5dbd9-jvp2s                1/1       Running   0          27s
slime-boot-68b6f88b7b-wwqnd             1/1       Running   0          39s
```

```
$ kubectl get po -n {{your_namespace}}
NAME                              READY     STATUS    RESTARTS   AGE
global-sidecar-785b58d4b4-fl8j4   1/1       Running   0          68s
```

3. 打开配置懒加载：
   业务namespace已有应用，在业务namespace中创建servicefence，执行`kubectl apply -f servicefence.yaml`

```yaml
apiVersion: microservice.slime.io/v1alpha1
kind: ServiceFence
metadata:
  name: {{your_svc}}
  namespace: {{you_namespace}}
spec:
  enable: true
```

4. 确认懒加载已开启
   执行`kubectl get sidecar {{your svc}} -oyaml`，可以看到对应服务生成了一个sidecar，如下：

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Sidecar
metadata:
  name: {{your_svc}}
  namespace: {{your_namespace}}
  ownerReferences:
  - apiVersion: microservice.slime.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: ServiceFence
    name: {{your_svc}}
spec:
  egress:
  - hosts:
    - istio-system/*
    - mesh-operator/*
    - '*/global-sidecar.{{your ns}}.svc.cluster.local'
  workloadSelector:
    labels:
      app: {{your_svc}}
```



### 使用集群唯一的global-sidecar   

> [完整样例](./install/samples/lazyload/slimeboot_lazyload_cluster_global_sidecar.yaml)
>
> 使用说明：
>
> k8s体系里，短域名访问的流量只会来自于同namespace，跨namespace访问必须带有namespace信息。cluster级别的global-sidecar和业务应用往往不在同namespace下，缺少短域名的配置，其拥有的配置必然带有namespace信息，因此global-sidecar无法成功转发同namespace内的访问请求，导致超时 "HTTP/1.1 0 DC downstream_remote_disconnect"错误。
>
> 所以，使用集群级global-sidecar时，应用间访问要携带namespace信息。

```yaml
apiVersion: config.netease.com/v1alpha1
kind: SlimeBoot
metadata:
  name: lazyload
  namespace: mesh-operator
spec:
  image:
    pullPolicy: Always
    repository: docker.io/slimeio/slime-lazyload
    tag: {{your_lazyload_tag}}
  module:
    - fence:
        enable: true
        wormholePort:
        - "{{your_port}}" # replace to your application service ports, and extend the list in case of multi ports
      name: slime-fence
      global:
        misc:
          global-sidecar-mode: cluster
      metric:
        prometheus:
          address: {{prometheus_address}} # replace to your prometheus address
          handlers:
            destination:
              query: |
                sum(istio_requests_total{source_app="$source_app",reporter="destination"})by(destination_service)
              type: Group
  component:
    globalSidecar:
      enable: true
      type: cluster
      image:
        repository: {{your_sidecar_repo}}
        tag: {{your_sidecar_tag}}
    pilot:
      enable: true
      image:
        repository: {{your_pilot_repo}}
        tag: {{your_pilot_tag}}
```



### 不使用global-sidecar组件  

在开启allow_any的网格中，可以不使用global-sidecar组件。使用如下配置：

> [完整样例](./install/samples/lazyload/slimeboot_lazyload_no_global_sidecar.yaml)
>
> 使用说明：
>
> 不使用global-sidecar组件可能会导致首次调用无法按照预先设定的路由规则进行，可能走到istio的默认兜底逻辑（一般是passthrough），从而倒回到原来的clusterIP访问服务，配置的virtualservice路由会暂时失效。
>
> 场景：
>
> 服务A访问服务B，但服务B的virtualservice会将访问服务B的请求转到服务C。由于没有global sidecar兜底，第一次请求会被istio透传，经PassthroughCluster到服务B。本来应该由服务C响应，变成服务B响应，出错。后面A的servicefence会添加上B，随即感知B的virtualservice将请求导向C，所以第一次之后的请求都会成功由C响应。

```yaml
apiVersion: config.netease.com/v1alpha1
kind: SlimeBoot
metadata:
  name: lazyload
  namespace: mesh-operator
spec:
  image:
    pullPolicy: Always
    repository: docker.io/slimeio/slime-lazyload
    tag: {{your_lazyload_tag}}
  module:
    - fence:
        enable: true
        wormholePort:
        - "{{your_port}}" # replace to your application service ports, and extend the list in case of multi ports
      name: slime-fence
      global:
        misc:
          global-sidecar-mode: no
      metric:
        prometheus:
          address: {{prometheus_address}} # replace to your prometheus address
          handlers:
            destination:
              query: |
                sum(istio_requests_total{source_app="$source_app",reporter="destination"})by(destination_service)
              type: Group
```

不使用global-sidecar组件可能会导致首次调用无法按照预先设定的路由规则进行。 





## 其他特性介绍

### 基于Accesslog开启懒加载

指定SlimeBoot CR资源中`spec.module.global.misc.metric_source_type`等于`accesslog`会使用Accesslog获取服务调用关系，等于`prometheus`则使用Prometheus。

使用Accesslog获取服务调用关系的大概过程：

- slime-boot在创建global-sidecar时，发现`metric_source_type: accesslog`，额外生成一个configmap，内容是包含lazyload controller处理accesslog的地址信息的static_resources。再通过一个envoyfilter，将static_resources加入global-sidecar配置中，使得global-sidecar的accesslog会发送到lazyload controller
- global-sidecar完成兜底转发时会生成accesslog，包含了调用方和被调用方服务信息。global-sidecar将信息发送给lazyload controller
- lazyload controller分析accesslog，获取到新的服务调用关系

随后的过程，就是修改servicefence和sidecar，和处理prometheus metric的过程一致。

样例

```yaml
spec:
  module:
    - name: lazyload
      enable: true
      fence:
        wormholePort: # replace to your application svc ports
          - "9080"
      global:
        misc:
          metric_source_type: accesslog
```

[完整样例](./install/samples/lazyload/slimeboot_lazyload_accesslog.yaml)



### 基于namespace/service label自动生成ServiceFence



fence支持基于label的自动生成，也即可以通过打label `slime.io/serviceFenced`的方式来定义**”开启fence“功能的范围**。

* namespace级别

  * `true`： 会对该namespace下的所有（没有cr的）服务都创建servicefence cr
  * 其他值： 无操作

* service级别

  * `true`： 对该服务生成servicefence cr
  * `false`： 不对该服务生成servicefence cr

  > 以上都会覆盖namespace级别设置（label）

  * 其他值： 使用namespace级别配置



对于自动生成的servicefence cr，会通过标准label `app.kubernetes.io/created-by=fence-controller`来记录，实现了状态关联变更。 而不匹配该label的servicefence，目前视为手动配置，不受以上label影响。



**举例**

> namespace `testns`下有三个服务： `svc1`, `svc2`, `svc3`

* 给`testns`打上`slime.io/serviceFenced=true` label： 生成以上三个服务的cr
* 给`svc2`打上 `slime.io/serviceFenced=false` label： 只剩下`svc1`, `svc3`这两个cr
* 删掉`svc2`的该label：恢复三个cr
* 去掉`svc3`的cr的`app.kubernetes.io/created-by=fence-controller`； 去掉`testns`上的label： 只剩下`svc3`的cr



**配置样例**



```yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    istio-injection: enabled
    slime.io/serviceFenced: "true"
  name: testns
---
apiVersion: v1
kind: Service
metadata:
  annotations: {}
  labels:
    app: svc2
    service: svc2
    slime.io/serviceFenced: "false"
  name: svc2
  namespace: testns
```





### 自定义兜底流量分派

lazyload/fence默认会将envoy无法匹配路由（缺省）的流量兜底发送到global sidecar，应对短暂服务数据缺失的问题，这是“懒加载”所必然面对的。 该方案因为技术细节上的局限性，对于目标（如域名）是集群外的流量，无法正常处理，详见 [[Configuration Lazy Loading]: Failed to access external service #3](https://github.com/slime-io/slime/issues/3)。

基于这个背景，设计了本特性，同时也能用于更灵活的业务场景。 大致思路是通过域名匹配的方式将不同的缺省流量分派到不同的目标做正确处理。



配置样例：

```yaml
module:
  - name: fence
    fence:
      wormholePort:
      - "80"
      - "8080"
      dispatches:  # new field
      - name: 163
        domains:
        - "www.163.com"
        cluster: "outbound|80||egress1.testns.svc.cluster.local"  # standard istio cluster format: <direction>|<svcPort>|<subset>|<svcFullName>, normally direction is outbound and subset is empty      
      - name: baidu
        domains:
        - "*.baidu.com"
        - "baidu.*"
        cluster: "{{ (print .Values.foo \".\" .Values.namespace ) }}"  # you can use template to construct cluster dynamically
      - name: sohu
        domains:
        - "*.sohu.com"
        - "sodu.*"
        cluster: "_GLOBAL_SIDECAR"  # a special name which will be replaced with actual global sidecar cluster
      - name: default
        domains:
        - "*"
        cluster: "PassthroughCluster"  # a special istio cluster which will passthrough the traffic according to orgDest info. It's the default behavior of native istio.

foo: bar
```

> 在本例中，我们把一部分流量分派给了指定的cluster； 另一部分让它走global sidecar； 然后对其余的流量，让它保持原生istio的行为： passthrough



**注意**：

* 自定义分派场景，如果希望保持原有逻辑 “其他所有未定义流量走global sidecar” 的话，需要显式配置如上的最后一条



### 日志输出到本地并轮转

slime的日志默认输出到标准输出，指定SlimeBoot CR资源中`spec.module.global.log.logRotate`等于`true`会将日志输出到本地并启动日志轮转，不再输出到标准输出。

轮转配置也是可调整的，默认的配置如下，可以通过显示指定logRotateConfig中的各个值进行覆盖。

```yaml
spec:
  module:
    - name: lazyload
      enable: true
      fence:
        wormholePort: # replace to your application svc ports
          - "9080"
      global:
        log:
          logRotate: true
          logRotateConfig:
            filePath: "/tmp/log/slime.log"
            maxSizeMB: 100
            maxBackups: 10
            maxAgeDay: 10
            compress: true
```

通常需要配合存储卷使用，在存储卷准备完毕后，指定SlimeBoot CR资源中的`spec.volumes`和`spec.containers.slime.volumeMounts`来显示将存储卷挂载到日志本地文件所在的路径。

以下是基于minikube kubernetes场景下的完整demo

#### 创建存储卷

基于/mnt/data路径创建hostpath类型的存储卷

```yaml
# hostPath pv for minikube demo
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: lazyload-claim
  namespace: mesh-operator
spec:
  storageClassName: manual
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 3Gi
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: lazyload-volumn
  labels:
    type: local
spec:
  storageClassName: manual
  capacity:
    storage: 5Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data"
```

#### 在SlimeBoot中声明挂载信息

在SlimeBoot CR资源中指定了存储卷会挂载到pod的"/tmp/log"路径，这样slime的日志会持久化到/mnt/data路径下，并且会自动轮转。

```yaml
---
apiVersion: config.netease.com/v1alpha1
kind: SlimeBoot
metadata:
  name: lazyload
  namespace: mesh-operator
spec:
  image:
    pullPolicy: Always
    repository: docker.io/slimeio/slime-lazyload
    tag: master-e5f2d83-dirty_1b68486
  module:
    - name: lazyload
      enable: true
      fence:
        wormholePort: # replace to your application svc ports
          - "9080"
      global:
        log:
          logRotate: true
          logRotateConfig:
            filePath: "/tmp/log/slime.log"
            maxSizeMB: 100
            maxBackups: 10
            maxAgeDay: 10
            compress: true
#...
  volumes:
    - name: lazyload-storage
      persistentVolumeClaim:
        claimName: lazyload-claim
  containers:
    slime:
      volumeMounts:
        - mountPath: "/tmp/log"
          name: lazyload-storage
```

[完整样例](./install/samples/lazyload/slimeboot_lazyload_logrotate.yaml)





## 示例

为bookinfo的productpage服务开启懒加载

### 安装 istio (1.8+)



### 设定tag

$latest_tag获取最新tag。默认执行的shell脚本和yaml文件均是$latest_tag版本。

```sh
$ export latest_tag=$(curl -s https://api.github.com/repos/slime-io/lazyload/tags | grep 'name' | cut -d\" -f4 | head -1)
```



### 安装 slime 

```shell
$ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/slime-io/lazyload/$latest_tag/install/samples/lazyload/easy_install_lazyload.sh)"
```

确认所有组件已正常运行

```sh
$ kubectl get slimeboot -n mesh-operator
NAME       AGE
lazyload   2m20s
$ kubectl get pod -n mesh-operator
NAME                                    READY   STATUS             RESTARTS   AGE
global-sidecar-pilot-7bfcdc55f6-977k2   1/1     Running            0          2m25s
lazyload-b9646bbc4-ml5dr                1/1     Running            0          2m25s
slime-boot-7b474c6d47-n4c9k             1/1     Running            0          4m55s
$ kubectl get po -n default
NAME                              READY   STATUS    RESTARTS   AGE
global-sidecar-59f4c5f989-ccjjg   1/1     Running   0          3m9s
```



### 安装bookinfo

   创建前请将current-context中namespace切换到你想部署bookinfo的namespace，使bookinfo创建在其中。此处以default为例。

```sh
$ kubectl label namespace default istio-injection=enabled
$ kubectl apply -f "https://raw.githubusercontent.com/slime-io/lazyload/$latest_tag/install/config/bookinfo.yaml"
```

创建完后，状态如下

```sh
$ kubectl get po -n default
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-79f774bdb9-6vzj6       2/2     Running   0          60s
global-sidecar-59f4c5f989-ccjjg   1/1     Running   0          5m12s
productpage-v1-6b746f74dc-vkfr7   2/2     Running   0          59s
ratings-v1-b6994bb9-klg48         2/2     Running   0          59s
reviews-v1-545db77b95-z5ql9       2/2     Running   0          59s
reviews-v2-7bf8c9648f-xcvd6       2/2     Running   0          60s
reviews-v3-84779c7bbc-gb52x       2/2     Running   0          60s
```

此样例中可以在pod/ratings中发起对productpage的访问，`curl productpage:9080/productpage`。

另外也可参考 [对外开放应用程序](https://istio.io/latest/zh/docs/setup/getting-started/#ip) 给应用暴露外访接口。



### 开启懒加载

创建servicefence，为productpage服务启用懒加载。

```sh
$ kubectl apply -f "https://raw.githubusercontent.com/slime-io/lazyload/$latest_tag/install/samples/lazyload/servicefence_productpage.yaml"
```

确认生成servicefence和sidecar对象。

```sh
$ kubectl get servicefence -n default
NAME          AGE
productpage   12s
$ kubectl get sidecar -n default
NAME          AGE
productpage   22s
$ kubectl get sidecar productpage -n default -oyaml
apiVersion: networking.istio.io/v1beta1
kind: Sidecar
metadata:
  creationTimestamp: "2021-08-04T03:54:35Z"
  generation: 1
  name: productpage
  namespace: default
  ownerReferences:
  - apiVersion: microservice.slime.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: ServiceFence
    name: productpage
    uid: d36e4be7-d66c-4f77-a9ff-14a4bf4641e6
  resourceVersion: "324118"
  uid: ec283a14-8746-42d3-87d1-0ee4538f0ac0
spec:
  egress:
  - hosts:
    - istio-system/*
    - mesh-operator/*
    - '*/global-sidecar.default.svc.cluster.local'
  workloadSelector:
    labels:
      app: productpage
```


### 首次访问观察

第一次访问productpage，并使用`kubectl logs -f productpage-xxx -c istio-proxy -n default`观察访问日志。

```
[2021-08-06T06:04:36.912Z] "GET /details/0 HTTP/1.1" 200 - via_upstream - "-" 0 178 43 43 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36" "48257260-1f5f-92fa-a18f-ff8e2b128487" "details:9080" "172.17.0.17:9080" outbound|9080||global-sidecar.default.svc.cluster.local 172.17.0.11:45422 10.101.207.55:9080 172.17.0.11:56376 - -
[2021-08-06T06:04:36.992Z] "GET /reviews/0 HTTP/1.1" 200 - via_upstream - "-" 0 375 1342 1342 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36" "48257260-1f5f-92fa-a18f-ff8e2b128487" "reviews:9080" "172.17.0.17:9080" outbound|9080||global-sidecar.default.svc.cluster.local 172.17.0.11:45428 10.106.126.147:9080 172.17.0.11:41130 - -
```

可以看出，此次outbound后端访问global-sidecar.default.svc.cluster.local。

观察sidecar内容

```sh
$ kubectl get sidecar productpage -oyaml
apiVersion: networking.istio.io/v1beta1
kind: Sidecar
metadata:
  creationTimestamp: "2021-08-06T03:23:05Z"
  generation: 2
  name: productpage
  namespace: default
  ownerReferences:
  - apiVersion: microservice.slime.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: ServiceFence
    name: productpage
    uid: 27853fe0-01b3-418f-a785-6e49db0d201a
  resourceVersion: "498810"
  uid: e923e426-f0f0-429a-a447-c6102f334904
spec:
  egress:
  - hosts:
    - '*/details.default.svc.cluster.local'
    - '*/reviews.default.svc.cluster.local'
    - istio-system/*
    - mesh-operator/*
    - '*/global-sidecar.default.svc.cluster.local'
  workloadSelector:
    labels:
      app: productpage
```

reviews 和 details 被自动加入！



### 再次访问观察

第二次访问productpage，观察productpage应用日志

```
[2021-08-06T06:05:47.068Z] "GET /details/0 HTTP/1.1" 200 - via_upstream - "-" 0 178 46 46 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36" "1c1c8e23-24d3-956e-aec0-e4bcff8df251" "details:9080" "172.17.0.6:9080" outbound|9080||details.default.svc.cluster.local 172.17.0.11:58522 10.101.207.55:9080 172.17.0.11:57528 - default
[2021-08-06T06:05:47.160Z] "GET /reviews/0 HTTP/1.1" 200 - via_upstream - "-" 0 379 1559 1558 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36" "1c1c8e23-24d3-956e-aec0-e4bcff8df251" "reviews:9080" "172.17.0.10:9080" outbound|9080||reviews.default.svc.cluster.local 172.17.0.11:60104 10.106.126.147:9080 172.17.0.11:42280 - default
```

可以看到，outbound日志的后端访问信息变为details.default.svc.cluster.local和reviews.default.svc.cluster.local。



### 卸载

卸载bookinfo

```sh
$ kubectl delete -f "https://raw.githubusercontent.com/slime-io/lazyload/$latest_tag/install/config/bookinfo.yaml"
```

卸载slime相关

```sh
$ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/slime-io/lazyload/$latest_tag/install/samples/lazyload/easy_uninstall_lazyload.sh)"
```



### 补充说明

如想要使用其他tag或commit_id的shell脚本和yaml文件，请显示指定$custom_tag_or_commit。

```sh
$ export custom_tag_or_commit=xxx
```

执行的命令涉及到yaml文件，用$custom_tag_or_commit替换$latest_tag，如下

```sh
#$ kubectl apply -f "https://raw.githubusercontent.com/slime-io/lazyload/$latest_tag/install/config/bookinfo.yaml"
$ kubectl apply -f "https://raw.githubusercontent.com/slime-io/lazyload/$custom_tag_or_commit/install/config/bookinfo.yaml"
```

执行的命令涉及到shell文件，将$custom_tag_or_commit作为shell文件的参数，如下

```sh
#$ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/slime-io/lazyload/$latest_tag/install/samples/smartlimiter/easy_install_limiter.sh)"
$ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/slime-io/lazyload/$latest_tag/install/samples/smartlimiter/easy_install_limiter.sh)" $custom_tag_or_commit
```





## FAQ

### global-sidecar-pilot的意义？

由于global-sidecar的作用不同于普通sidecar，需要一些定制逻辑，比如兜底envoyfilter不对global-sidecar生效否则会死循环等，global-sidecar并不能直接使用集群原有pilot的全量配置。global-sidecar-pilot会从集群原有pilot处获取全量配置后，会进行微调，再推送给global-sidecar。现有global-sidecar-pilot是基于istiod 1.7改造的。

注：为了降低学习成本、增强兼容性，我们正在考虑去除global-sidecar-pilot，届时不再有定制化的pilot，完全兼容社区版本，预计在下一个大版本中实现此功能。



### global sidecar不能正常启动？

global sidecar启动报错`Internal:Error adding/updating listener(s) 0.0.0.0_15021: cannot bind '0.0.0.0:15021': Address already in use`，这错误通常是端口冲突导致。global-sidecar是以gateway模式运行的sidecar，它会绑定到真实端口上。具体来说是istio-ingressgateway使用了15021端口，这会导致global-sidecar的lds更新失败，修改ingressgateway的15021端口为其他值可解决。

注：目前是通过端口规划来解决此问题，下个大版本中会摆脱这个局限。
