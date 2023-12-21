package integration

var BusyboxImage = "registry.k8s.io/e2e-test-images/busybox:1.29-2"

func officialImages(names ...string) map[string]string {
	// basic mapping for now since there are
	// very few official Windows-based images
	m := map[string]string{
		"busybox:latest": "registry.k8s.io/e2e-test-images/busybox:1.29-2",
		"busybox":        "registry.k8s.io/e2e-test-images/busybox:1.29-2",
	}
	return m
}
