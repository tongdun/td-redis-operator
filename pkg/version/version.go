// Package version defines operator version
package version

var (
	// version from git
	gitVersion = "v0.0.0-master+$Format:%h$"
	// sha1 from git, output of $(git rev-parse HEAD)
	gitCommit = "$Format:%H$"
	// state of git tree, either "clean" or "dirty"
	gitTreeState = ""
	// build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	buildDate = "1970-01-01T00:00:00Z"
)

// Info defines version info of this project
type Info struct {
	BuildDate    string
	GitCommit    string
	GitTreeState string
	GitVersion   string
}

// String implements Stringer
func (v Info) String() string {
	return v.GitVersion
}

// Version returns current version info
func Version() Info {
	return Info{
		BuildDate:    buildDate,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GitVersion:   gitVersion,
	}
}
