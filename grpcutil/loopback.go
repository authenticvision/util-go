package grpcutil

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Loopback struct {
	URL      string
	Listener net.Listener
	sockPath string
}

func (l *Loopback) Close() error {
	if l.Listener != nil {
		if err := l.Listener.Close(); err != nil {
			return fmt.Errorf("close grpc loopback listener: %w", err)
		}
	}
	if l.sockPath != "" {
		if err := os.Remove(l.sockPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove grpc loopback socket: %w", err)
		}
	}
	return nil
}

func NewLoopback() (*Loopback, error) {
	// A temporary local unix socket will be spawned for grpc-gateway connectivity. Oh, well.
	// On Linux, this uses an abstract unix socket to avoid a dependency on a writeable filesystem.
	var result Loopback
	var sockPath string
	sockPrefix := "grpcloop"
	if exeName, err := os.Executable(); err == nil && exeName != "" {
		sockPrefix = filepath.Base(exeName)
		sockPrefix = strings.TrimLeft(sockPrefix, "_- ")
		sockPrefix = strings.ReplaceAll(sockPrefix, "_", "-") // cosmetics for go run
	}
	sockName := fmt.Sprintf("%v-%d.sock", sockPrefix, os.Getpid())
	if runtime.GOOS == "linux" {
		sockPath = "@" + sockName
		result.URL = "unix-abstract:" + sockName
	} else {
		sockPath = filepath.Join(os.TempDir(), sockName)
		result.URL = "unix://" + sockPath
		if err := os.Remove(sockPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("remove old socket: %w", err)
		}
		result.sockPath = sockPath // delete again on close
	}

	l, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("create socket: %w", err)
	}
	result.Listener = l

	return &result, nil
}
