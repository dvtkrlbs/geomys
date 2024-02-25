package geomys

import "strings"

type Module struct {
	Path    string // module path
	Version string
	Deps    []string
	Libs    []Package
}

type Package struct {
	Path    string
	Files   []string
	Imports []string
}

func (m *Module) AddDep(dep string) {
	m.Deps = append(m.Deps, dep)
}

func (m *Module) CanonicalName() string {
	lowered := strings.ToLower(m.Path)
	return CanonicalizeModuleName(lowered)
}

func (m *Module) AddLib(lib Package) {
	m.Libs = append(m.Libs, lib)
}

func (l *Package) CanonicalName() string {
	lowered := strings.ToLower(l.Path)
	return CanonicalizeModuleName(lowered)
}
