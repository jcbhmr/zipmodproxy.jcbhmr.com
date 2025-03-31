package modproxy

import (
	"io"
	"time"
)

type Info struct {
	Version string    // version string
	Time    time.Time // commit time
}

type ModProxier interface {
	List(modulePath string) ([]string, error)
	Info(modulePath, version string) (Info, error)
	Mod(modulePath, version string) (io.Reader, error)
	Zip(modulePath, version string) (io.Reader, error)
}

type ModProxierLatest interface {
	ModProxier
	Latest(modulePath string) (string, error)
}
