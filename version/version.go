package version

import (
	"runtime"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/xerror"
)

var BuildTime = ""
var Version = ""
var GoVersion = runtime.Version()
var GoPath = ""
var GoROOT = ""
var CommitID = ""
var Project = ""

func init() {
	if Version == "" {
		Version = "v0.0.1"
	}

	xerror.ExitErr(ver.NewVersion(Version))
}
