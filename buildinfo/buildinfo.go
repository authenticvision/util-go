package buildinfo

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

var (
	GitCommit     string
	GitCommitDate *time.Time
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}

	commit := ""
	dirty := false
	ts := time.Time{}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			commit = setting.Value
		} else if setting.Key == "vcs.time" {
			ts, _ = time.Parse(time.RFC3339, setting.Value)
		} else if setting.Key == "vcs.modified" {
			dirty, _ = strconv.ParseBool(setting.Value)
		}
	}
	GitCommit = commit
	if dirty {
		GitCommit += "-dirty"
	}
	GitCommitDate = &ts
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(w, "commit: %v\ncommit date: %v\n",
		GitCommit, GitCommitDate)
})

func Print(name string) {
	fmt.Printf("%s  commit: %s (%v)\n", name, GitCommit, GitCommitDate)
}
