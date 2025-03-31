package main

import (
	"bytes"
	"io"
	"strings"

	"github.com/jcbhmr/zipmodproxy.jcbhmr.com/internal/modproxy"
)

type MyModProxier struct{}

func (m *MyModProxier) List(modulePath string) ([]string, error) {
	return []string{}, nil
}

func (m *MyModProxier) Info(modulePath, version string) (modproxy.Info, error) {
	return modproxy.Info{}, nil
}

func (m *MyModProxier) Mod(modulePath, version string) (io.Reader, error) {
	return strings.NewReader(""), nil
}

func (m *MyModProxier) Zip(modulePath, version string) (io.Reader, error) {
	return bytes.NewReader([]byte("")), nil
}

func (m *MyModProxier) Latest(modulePath string) (string, error) {
	return "", nil
}

func main() {
	start(&modproxy.ModProxyMux{ModProxier: &MyModProxier{}})
}
