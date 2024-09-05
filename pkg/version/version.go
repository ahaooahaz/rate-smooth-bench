package version

import "fmt"

// Version information.
var (
	BuildTS   = "None"
	GitHash   = "None"
	GitBranch = "None"
	Version   = "None"
	App       = "None"
)

// GetApp ...
func GetApp() string {
	if App != "None" {
		return fmt.Sprintf("%s-%s", App, GitBranch)
	}
	return App
}

// GetVersion Printer print build version
func GetVersion() string {
	if GitHash != "None" {
		h := GitHash
		if len(h) > 7 {
			h = h[:7]
		}
		return fmt.Sprintf("%s-%s", Version, h)
	}
	return Version
}

func GetFullVersionInfo() string {
	f := fmt.Sprintf("Application:      %v\n", App)
	f += fmt.Sprintf("Version:          %v\n", GetVersion())
	f += fmt.Sprintf("Git Branch:       %v\n", GitBranch)
	f += fmt.Sprintf("Git Commit:       %v\n", GitHash)
	f += fmt.Sprintf("Build Time:       %v\n", BuildTS)
	return f
}
