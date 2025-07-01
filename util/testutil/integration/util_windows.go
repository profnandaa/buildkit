package integration

import (
	"golang.org/x/sys/windows/registry"
	"net"
	"regexp"

	"github.com/Microsoft/go-winio"

	// include npipe connhelper for windows tests
	_ "github.com/moby/buildkit/client/connhelper/npipe"
)

var socketScheme = "npipe://"

var windowsImagesMirrorMap = map[string]string{
	// TODO(profnandaa): currently, amd64 only, to revisit for other archs.
	"nanoserver:latest": "mcr.microsoft.com/windows/nanoserver:ltsc" + getWinServerYear(),
	"servercore:latest": "mcr.microsoft.com/windows/servercore:ltsc" + getWinServerYear(),
	"busybox:latest":    "registry.k8s.io/e2e-test-images/busybox@sha256:6d854ffad9666d2041b879a1c128c9922d77faced7745ad676639b07111ab650",
	// nanoserver with extra binaries, like fc.exe
	// TODO(profnandaa): get an approved/compliant repo, placeholder for now
	// see dockerfile here - https://github.com/microsoft/windows-container-tools/pull/178
	"nanoserver:plus":         "docker.io/wintools/nanoserver:ltsc2022",
	"nanoserver:plus-busybox": "docker.io/wintools/nanoserver:ltsc2022",
}

func getWinServerYear() string {
	fallback := "2022"
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return fallback
	}
	defer k.Close()

	productName, _, err := k.GetStringValue("ProductName")
	if err != nil {
		return fallback
	}

	re := regexp.MustCompile(`\b(20\d{2})`)
	return re.FindString(productName)
}

// abstracted function to handle pipe dialing on windows.
// some simplification has been made to discard timeout param.
func dialPipe(address string) (net.Conn, error) {
	return winio.DialPipe(address, nil)
}
