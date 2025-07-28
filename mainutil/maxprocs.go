//go:build !go1.25

package mainutil

import "go.uber.org/automaxprocs/maxprocs"

func autoMaxProcs() error {
	_, err := maxprocs.Set()
	return err
}
