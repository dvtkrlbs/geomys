package geomys

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

type Module struct {
	Path       string       // module path
	Query      string       // version query corresponding to this version
	Version    string       // module version
	Versions   []string     // available module versions
	Replace    *Module      // replaced by this module
	Time       *time.Time   // time version was created
	Update     *Module      // available update (with -u)
	Main       bool         // is this the main module?
	Indirect   bool         // module is only indirectly needed by main module
	Dir        string       // directory holding local copy of files, if any
	GoMod      string       // path to go.mod file describing module, if any
	GoVersion  string       // go version used in module
	Retracted  []string     // retraction information, if any (with -retracted or -u)
	Deprecated string       // deprecation message, if any (with -u)
	Error      *ModuleError // error loading module
	Origin     any          // provenance of module
	Reuse      bool         // reuse of old module info is safe
	Deps       []string
}

type ModuleError struct {
	Err string // the error itself
}

func DepList() ([]Module, error) {
	out, err := exec.Command("go", "list", "-m", "-json", "all").Output()
	if err != nil {
		return nil, fmt.Errorf("error runing go list: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("error running go list: %w", err)
	}

	modules := make([]Module, 0)
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var module Module

		err := dec.Decode(&module)
		if err == io.EOF {
			// all done
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error decoding module: %w", err)
		}

		modules = append(modules, module)
	}

	return modules, nil
}

func DepGraph() (map[string][]string, error) {
	out, err := exec.Command("go", "mod", "graph").Output()
	if err != nil {
		return nil, fmt.Errorf("error running go mod graph: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))

	moduleMap := make(map[string][]string)

	for scanner.Scan() {
		line := scanner.Text()

		separatedLine := strings.Split(line, " ")
		key, value := separatedLine[0], separatedLine[1]

		//separatedKey := strings.Split(key, "@")
		//keyName, _ := separatedKey[0], separatedKey[1]

		//separatedValue := strings.Split(value, "@")
		//valueName, _ := separatedValue[0], separatedValue[1]

		if strings.HasPrefix(key, "go@") || strings.HasPrefix(value, "go@") {
			continue
		}
		mod, ok := moduleMap[key]

		if ok {
			moduleMap[key] = append(mod, value)
		} else {
			moduleMap[key] = []string{value}
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input %v", err)
	}

	return moduleMap, nil
}
