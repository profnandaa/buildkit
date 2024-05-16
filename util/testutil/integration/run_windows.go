package integration

var BusyboxImage = "registry.k8s.io/e2e-test-images/busybox:1.29-2"

func officialImages(_ ...string) map[string]string {
	// supplied string will be ignored, intended for UNIX.
	// basic mapping for now since there are
	// very few official Windows-based images.
	m := map[string]string{
		"library/nanoserver:latest": "mcr.microsoft.com/windows/nanoserver:ltsc2022",
		"library/servercore:latest": "mcr.microsoft.com/windows/servercore:ltsc2022",
		"library/busybox:latest":    "registry.k8s.io/e2e-test-images/busybox:1.29-2",
	}
	return m
}
