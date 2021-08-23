package version

import (
	"fmt"
)

// Build information. Populated at build-time.
var (
	Version = "dev"
	Build   string
)

// versionInfoTmpl contains the template used by Print.
var versionInfoTmpl = `%s, version %s (build: %s)`

// Print returns version information.
func Print(program string) string {
	return fmt.Sprintf(versionInfoTmpl, program, Version, Build)
}
