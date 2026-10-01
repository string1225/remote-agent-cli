package buildinfo

import "runtime/debug"

// Version is set by release builds, including builds from source archives.
var Version = "dev"

func Current() string {
	if Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		revision, dirty := "", false
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			case "vcs.modified":
				dirty = setting.Value == "true"
			}
		}
		if len(revision) >= 7 {
			if dirty {
				return revision[:7] + "-dirty"
			}
			return revision[:7]
		}
	}
	return Version
}
