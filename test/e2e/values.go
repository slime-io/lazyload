package e2e

var (
	testResourceToDelete  []*TestResource
	nsSlime               = "mesh-operator"
	nsApps                = "example-apps"
	test                  = "test/e2e/testdata/install"
	slimebootName         = "slime-boot"
	istioRevKey           = "istio.io/rev"
	istioRevValue         = "1-10-2"
	slimebootTag          = "v0.2.3-5bf313f"
	lazyloadTag           = "master-2946c4a"
	globalSidecarTag      = "1.7.0"
	globalSidecarPilotTag = "globalPilot-7.0-v0.0.3-713c611962"
	svfGroup = "microservice.slime.io"
	svfVersion = "v1alpha1"
	svfResource = "servicefences"
	svfName = "productpage"
	sidecarGroup = "networking.istio.io"
	sidecarVersion = "v1beta1"
	sidecarResource = "sidecars"
	sidecarName = "productpage"
)

type TestResource struct {
	Namespace string
	Contents  string
	Selectors []string
}
