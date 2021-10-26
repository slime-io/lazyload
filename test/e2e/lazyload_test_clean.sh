#!/bin/bash
sed -e 's/{{istioRevKey}}/istio.io\/rev/g' -e 's/{{istioRevValue}}/1-10-2/g' ./testdata/install/samples/lazyload/servicefence_productpage.yaml > tmp_servicefence_productpage.yaml
kubectl delete -f tmp_servicefence_productpage.yaml

sed -e 's/{{lazyloadTag}}/v0.2.6-d808438/g' -e 's/{{istioRevValue}}/1-10-2/g' -e 's/{{globalSidecarTag}}/1.7.0/g' -e 's/{{globalSidecarPilotTag}}/globalPilot-7.0-v0.0.3-713c611962/g' ./testdata/install/samples/lazyload/slimeboot_lazyload.yaml > tmp_slimeboot_lazyload.yaml
kubectl delete -f tmp_slimeboot_lazyload.yaml

kubectl delete -f ./testdata/install/config/bookinfo.yaml

sed -e 's/{{slimebootTag}}/v0.2.3-5bf313f/g' ./testdata/install/init/deployment_slime-boot.yaml > tmp_deployment_slime-boot.yaml
kubectl delete -f tmp_deployment_slime-boot.yaml

kubectl delete -f ./testdata/install/init/crds.yaml

kubectl delete ns example-apps

kubectl delete ns mesh-operator

rm -f tmp_*