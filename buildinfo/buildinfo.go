package buildinfo

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

var (
	// GitCommit is set via, or derived from debug VCS info:
	// -X github.com/authenticvision/util-go/buildinfo.GitCommit=${GIT_COMMIT}
	GitCommit string

	// gitCommitUnixTS is set via:
	// -X github.com/authenticvision/util-go/buildinfo.gitCommitUnixTS=${GIT_COMMIT_UNIXTIME}
	gitCommitUnixTS string

	// GitCommitDate is derived from buildinfo.gitCommitUnixTS or debug VCS info
	GitCommitDate time.Time

	// Version is set via:
	// -X github.com/authenticvision/util-go/buildinfo.Version=${GIT_VERSION}
	Version string
)

func init() {
	if gitCommitUnixTS != "" {
		i, err := strconv.ParseInt(gitCommitUnixTS, 10, 64)
		if err != nil {
			panic("error parsing git commit unix timestamp: " + err.Error())
		}
		GitCommitDate = time.Unix(i, 0)
	}

	if GitCommit == "" || GitCommitDate.IsZero() {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					GitCommit = setting.Value
				} else if setting.Key == "vcs.time" {
					var err error
					GitCommitDate, err = time.Parse(time.RFC3339, setting.Value)
					if err != nil {
						panic("error parsing git commit date: " + err.Error())
					}
				}
			}
		}
	}
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	type response struct {
		GitCommit     string    `json:"git_commit"`
		GitCommitDate time.Time `json:"git_commit_date"`
		Version       string    `json:"version,omitempty"`
	}
	resp := response{
		GitCommit:     GitCommit,
		GitCommitDate: GitCommitDate,
		Version:       Version,
	}
	w.Header().Add("Content-Type", "application/json")
	jw := json.NewEncoder(w)
	jw.SetIndent("", "  ")
	_ = jw.Encode(resp)
})
