package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/bazelbuild/bazel-gazelle/merger"
	"github.com/bazelbuild/bazel-gazelle/rule"
	bzl "github.com/bazelbuild/buildtools/build"
	"github.com/dvtkrlbs/geomys"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var testFile = `
http_archive(
    name = "anstyle-1.0.6.crate",
    sha256 = "8901269c6307e8d93993578286ac0edf7f195079ffff5ebdeea6a59ffb7e36bc",
    strip_prefix = "anstyle-1.0.6",
    urls = ["https://crates.io/api/v1/crates/anstyle/1.0.6/download"],
	test = glob(["**/*_test.go"]),
    visibility = [],
)
`

func main() {
	//_, err := geomys.DepList()
	//if err != nil {
	//	log.Fatalf("error getting dependency list: %v\n", err)
	//}

	graph, err := geomys.DepGraph()
	if err != nil {
		log.Fatalf("error getting dependency graph: %v\n", err)
	}

	// for _, module := range modules {
	// fmt.Printf("%+v\n", module)
	// }

	// fmt.Printf("%+v\n", graph)

	client := &http.Client{}

	file, err := os.OpenFile("third-party/BUCK", os.O_RDWR, 0644)
	if err != nil {
		newFile, err := os.Create("third-party/BUCK")
		if err != nil {
			log.Fatalf("error creating BUCK file: %v\n", err)
		}

		file = newFile
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("error reading BUCK file: %v\n", err)
	}
	bzlFile, err := bzl.ParseBuild("BUCK", fileBytes)
	if err != nil {
		log.Fatalf("error parsing BUCK file: %v\n", err)
	}

	ast := rule.ScanAST("geomys", bzlFile)
	// fmt.Printf("%+v\n", file)

	var rules []*rule.Rule
	seenModules := make(map[string]bool)
	modules := make([]geomys.Module, 0)
	rulesIndex := 0
	for module, deps := range graph {
		module := geomys.Module{
			Path: module,
		}
		for _, dep := range deps {
			if seenModules[dep] {
				continue
			}

			seenModules[dep] = true
			split := strings.Split(dep, "@")
			path, version := split[0], split[1]
			module := geomys.Module{
				Path:    path,
				Version: version,
			}

			nName := strings.ToLower(path)
			canonName := geomys.CanonicalizeModuleName(nName)
			//tempFile, _ := os.CreateTemp("geomys", fmt.Sprintf("%s@%s", canonName, dep.Version))

			body, err := geomys.GetModule(nName, version, client)
			if err != nil {
				log.Fatalf("error getting module: %v\n", err)
			}

			zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
			if err != nil {
				log.Fatalf("error reading zip: %v\n", err)
			}

			fileContents := make([]string, 0)
			importMaps := make(map[string]map[string]bool)
			for _, file := range zipReader.File {
				//fmt.Println(file.Name)

				if strings.HasSuffix(file.Name, ".go") && !strings.HasSuffix(file.Name, "_test.go") && !strings.Contains(file.Name, "testdata") {
					stripped := strings.TrimPrefix(file.Name, fmt.Sprintf("%s@%s/", nName, version))
					//if !strings.Contains(stripped, "/") {
					if strings.Contains(stripped, "@") || strings.HasPrefix(stripped, "cmd") || strings.Contains(stripped, "example") {
						continue
					}

					fileContent, err := file.Open()
					if err != nil {
						log.Fatalf("error opening file: %v\n", err)
					}
					imports, err := geomys.GetImports(fileContent)
					if err != nil {
						log.Fatalf("error getting imports for %s: %v\n", file.Name, err)
					}
					for _, imp := range imports {
						basePath := strings.TrimSuffix(filepath.Dir(file.Name), "/")
						basePath = strings.TrimPrefix(basePath, dep)
						basePath = filepath.Join(path, basePath)

						if importMaps[basePath] == nil {
							importMaps[basePath] = make(map[string]bool)
						}
						importMaps[basePath][imp] = true

					}
					fileContents = append(fileContents, stripped)
					//}
				}
			}

			sum := sha256.Sum256(body)

			archiveRule := rule.NewRule("http_archive", fmt.Sprintf("%s@%s.mod", canonName, version))

			archiveRule.SetAttr("strip_prefix", fmt.Sprintf("%s@%s", nName, version))
			archiveRule.SetAttr("visibility", []string{"PUBLIC"})
			archiveRule.SetAttr("sha256", fmt.Sprintf("%x", sum))
			archiveRule.SetAttr("sub_targets", fileContents)
			archiveRule.SetAttr("urls", []string{fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.zip", nName, version)})

			depsOfCurrent := graph[fmt.Sprintf("%s@%s", path, version)]
			goLibraryRule := rule.NewRule("go_library", fmt.Sprintf("%s@%s", canonName, version))

			deps := make([]string, 0)
			for dep, _ := range importMaps[path] {
				currentDep := ""
				for _, needle := range depsOfCurrent {
					if strings.HasPrefix(needle, dep) {
						currentDep = dep
						break
					}
				}
				if currentDep == "" {
					continue
				}
				//depName := strings.ToLower(dep)
				deps = append(deps, fmt.Sprintf("//third-party:%s", currentDep))
			}

			srcs := make([]string, 0)
			subLibs := make(map[string][]string)

			goLibraryRule.SetAttr("deps", deps)
			goLibraryRule.SetAttr("package_name", path)
			goLibraryRule.SetAttr("visibility", []string{"PUBLIC"})
			for _, file := range fileContents {
				dir, file := filepath.Split(file)
				dir = strings.TrimSuffix(dir, "/")
				if dir != "" {
					subLib, ok := subLibs[dir]
					if ok {
						subLibs[dir] = append(subLib, file)
					} else {
						subLibs[dir] = []string{file}
					}
				} else {
					srcs = append(srcs, fmt.Sprintf(":%s@%s.mod[%s]", canonName, version, file))
				}
			}
			goLibraryRule.SetAttr("srcs", srcs)

			for dir, files := range subLibs {
				subLibSrcs := make([]string, 0)
				cleanDir := strings.TrimSuffix(dir, "/")
				cleanDir = strings.ReplaceAll(cleanDir, "/", "_")
				subLibName := fmt.Sprintf("%s_%s@%s", canonName, cleanDir, version)
				subLibRule := rule.NewRule("go_library", subLibName)
				for _, file := range files {
					//filePath := geomys.CanonicalizeModuleName(filepath.Join(dir, file))
					filePath := filepath.Join(dir, file)
					subLibSrcs = append(subLibSrcs, fmt.Sprintf(":%s@%s.mod[%s]", canonName, version, filePath))
				}
				subLibRule.SetAttr("srcs", subLibSrcs)
				subLibDeps := make([]string, 0)
				for dep, _ := range importMaps[filepath.Join(path, dir)] {
					depDir := strings.TrimSuffix(dep, "/")
					depDir = strings.ReplaceAll(depDir, "/", "_")

					depName := fmt.Sprintf("%s@%s", canonName, version)
					subLibDeps = append(subLibDeps, fmt.Sprintf(":%s", depName))
				}
				//subLibRule.SetAttr("deps", []string{fmt.Sprintf(":%s@%s", canonName, version)})
				subLibRule.SetAttr("package_name", filepath.Join(path, dir))
				subLibRule.SetAttr("visibility", []string{"PUBLIC"})
				subLibRule.SetAttr("deps", subLibDeps)
				rules = append(rules, subLibRule)
				srcs = append(srcs, fmt.Sprintf(":%s", subLibName))
			}

			//glob := rule.GlobValue{
			//	Patterns: []string{fmt.Sprintf("$(location :%s@%s.mod)/*.go", canonName, dep.Version)},
			//	Excludes: []string{fmt.Sprintf("$(location :%s@%s.mod)/*_test.go", canonName, dep.Version)},
			//}
			rules = append(rules, archiveRule)
			rules = append(rules, goLibraryRule)
			rulesIndex++
		}
	}
	//newRule := rule.NewRule("http_archive", "anstyle-1.0.7.crate")
	kinds := make(map[string]rule.KindInfo, 0)

	//log.Printf("%v, %v, %v\n", ast, rules, kinds)

	// rule.MergeRules(file.Rules, []*rule.Rule{newRule}, mergeables, "")
	merger.MergeFile(ast, []*rule.Rule{}, rules, merger.PostResolve, kinds)

	// for _, rule := range file.Rules {
	// fmt.Printf("%+v\n", rule
	// if rule.Name() == "anstyle-1.0.6.crate" {
	// rule.SetAttr("test", "test")
	// }
	// }

	formatted := ast.Format()

	_, err = file.Write(formatted)
	if err != nil {
		log.Fatalf("error writing to BUCK file: %v\n", err)
	}
	//fmt.Printf("%s\n", formatted)

	// newRule := rule.NewRule("http_archive", "anstyle-1.0.6.crate")
	// fmt.Printf("%+v\n", newRule)

}
