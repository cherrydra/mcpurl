package version

import (
	"fmt"
	"runtime/debug"
)

var (
	Prefix           = "mcpurl"
	Version   string = "dev"
	GoVersion string = "Go"
	Commit    string = "git"
	BuildTime string
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	GoVersion = info.GoVersion
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			Commit = s.Value
			continue
		}
		if s.Key == "vcs.time" {
			BuildTime = s.Value
		}
	}
}

func Short() string {
	return fmt.Sprintf("%s/%s-%s", Prefix, Version, Commit[:min(7, len(Commit))])
}

func Long() string {
	return fmt.Sprintf("%s/%s-%s", Prefix, Version, Commit)
}
