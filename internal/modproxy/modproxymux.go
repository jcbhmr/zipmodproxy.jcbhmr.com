package modproxy

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/mod/module"
)

type ModProxyMux struct {
	ModProxier
	initOnce sync.Once
	mux      http.ServeMux
	mux2     http.ServeMux
}

func (m *ModProxyMux) modulePathValue(r *http.Request) (string, error) {
	escapedRawModulePath := r.PathValue("escapedRawModulePath")
	if escapedRawModulePath == "" {
		return "", errors.New("no escapedRawModulePath")
	}
	rawModulePath, err := url.PathUnescape(escapedRawModulePath)
	if err != nil {
		return "", err
	}
	err = module.CheckPath(rawModulePath)
	if err != nil {
		return "", err
	}
	return rawModulePath, nil
}

func (m *ModProxyMux) versionValue(r *http.Request) (string, error) {
	escapedVersion := r.PathValue("escapedVersion")
	if escapedVersion == "" {
		return "", errors.New("no escapedVersion")
	}
	version, err := url.PathUnescape(escapedVersion)
	if err != nil {
		return "", err
	}
	return version, nil
}

func (m *ModProxyMux) rewrite(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/@")
	if len(parts) != 2 {
		http.Error(w, "no /@ token", http.StatusBadRequest)
		return
	}
	rawModulePath := parts[0][1:]
	var rawAtRoute string
	var rawAtData string
	var rawAtSuffix string
	{
		firstSlash := strings.IndexByte(parts[1], '/')
		if firstSlash == -1 {
			rawAtRoute = parts[1]
		} else {
			rawAtRoute = parts[1][:firstSlash]
			rawAtDataAndSuffix := parts[1][firstSlash+1:]
			lastDot := strings.LastIndexByte(rawAtDataAndSuffix, '.')
			if lastDot == -1 {
				rawAtData = rawAtDataAndSuffix
			} else {
				rawAtData = rawAtDataAndSuffix[:lastDot]
				rawAtSuffix = rawAtDataAndSuffix[lastDot+1:]
			}
		}
	}
	r.URL.Path = "/" + url.PathEscape(rawModulePath)
	r.URL.Path += "/" + rawAtRoute
	if rawAtData != "" {
		r.URL.Path += "/" + url.PathEscape(rawAtData)
		if rawAtSuffix != "" {
			r.URL.Path += "/" + rawAtSuffix
		}
	}
	m.mux2.ServeHTTP(w, r)
}

func (m *ModProxyMux) list(w http.ResponseWriter, r *http.Request) {
	modulePath, err := m.modulePathValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	versions, err := m.List(modulePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body := strings.Join(versions, "\n")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Write([]byte(body))
}

func (m *ModProxyMux) info(w http.ResponseWriter, r *http.Request) {
	version, err := m.versionValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	modulePath, err := m.modulePathValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	info, err := m.Info(modulePath, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Write(body)
}

func (m *ModProxyMux) mod(w http.ResponseWriter, r *http.Request) {
	version, err := m.versionValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	modulePath, err := m.modulePathValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := m.Mod(modulePath, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if body, ok := body.(io.ReadCloser); ok {
		defer body.Close()
	}
	if body, ok := body.(interface{ Len() int }); ok {
		w.Header().Set("Content-Length", strconv.Itoa(body.Len()))
	}
	_, err = io.Copy(w, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *ModProxyMux) zip(w http.ResponseWriter, r *http.Request) {
	version, err := m.versionValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	modulePath, err := m.modulePathValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := m.Zip(modulePath, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	if body, ok := body.(io.ReadCloser); ok {
		defer body.Close()
	}
	if body, ok := body.(interface{ Len() int }); ok {
		w.Header().Set("Content-Length", strconv.Itoa(body.Len()))
	}
	_, err = io.Copy(w, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *ModProxyMux) latest(w http.ResponseWriter, r *http.Request) {
	modProxier, ok := m.ModProxier.(ModProxierLatest)
	if !ok {
		http.NotFound(w, r)
		return
	}

	modulePath, err := m.modulePathValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	version, err := modProxier.Latest(modulePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(version)))
	w.Write([]byte(version))
}

func (m *ModProxyMux) init() {
	m.mux.HandleFunc("GET /{domainName}/...", m.rewrite)
	m.mux2.HandleFunc("GET /{escapedRawModulePath}/@v/list", m.list)
	m.mux2.HandleFunc("GET /{escapedRawModulePath}/@v/{escapedVersion}/info", m.info)
	m.mux2.HandleFunc("GET /{escapedRawModulePath}/@v/{escapedVersion}/mod", m.mod)
	m.mux2.HandleFunc("GET /{escapedRawModulePath}/@v/{escapedVersion}/zip", m.zip)
	m.mux2.HandleFunc("GET /{escapedRawModulePath}/@latest", m.latest)
}

func (m *ModProxyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.initOnce.Do(m.init)
	m.mux.ServeHTTP(w, r)
}
