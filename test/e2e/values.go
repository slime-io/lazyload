package e2e

var (
	testResourceToDelete []*TestResource
	nsSlime              = "mesh-operator"
	nsApps               = "example-apps"
	test                 = "test/e2e/testdata/install"
	slimebootName        = "slime-boot"
	istiodLabelKey       = "istio.io/rev"
	istiodLabelV         = "1-10-2"
	slimebootTag          = "v0.2.3-5bf313f"
	lazyloadTag           = "v0.2.6-d808438"
	globalSidecarTag      = "1.7.0"
	globalSidecarPilotTag = "globalPilot-7.0-v0.0.3-713c611962"
)

type TestResource struct {
	Namespace string
	Contents  string
	Selectors []string
}
