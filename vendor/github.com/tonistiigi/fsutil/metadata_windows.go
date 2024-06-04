//go:build windows
// +build windows

package fsutil

import "strings"

// there are some metadata files/directories that are not
// useful to containers and can be skipped
// they also require extra security privileges to read.
// lowercase for any future case changes
var MetadataFiles = map[string]bool{
	"\\system volume information":                                           true,
	"\\wcsandboxstate":                                                      true,
	"\\programdata\\microsoft\\diagnosis":                                   true,
	"\\program files\\windows defender advanced threat protection":          true,
	"\\programdata\\microsoft\\windows defender advanced threat protection": true,
	"\\windows\\system32\\logfiles\\wmi\\rtbackup":                          true,
	"\\windows\\globalization\\icu":                                         true,
}

func isMetadataFile(path string) bool {
	// normalize path
	if path[0] != '\\' {
		path = "\\" + path
	}
	path = strings.ToLower(path)
	return MetadataFiles[path]
}
