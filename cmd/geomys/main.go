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
	deps, err := geomys.DepList()
	if err != nil {
		log.Fatalf("error getting dependency list: %v\n", err)
	}

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
	rulesIndex := 0
	for _, dep := range deps {
		if dep.Main {
			continue

		}

		nName := strings.ToLower(dep.Path)
		canonName := geomys.CanonicalizeModuleName(nName)
		//tempFile, _ := os.CreateTemp("geomys", fmt.Sprintf("%s@%s", canonName, dep.Version))

		resp, err := client.Get(fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.zip", nName, dep.Version))
		if err != nil {
			log.Fatalf("error getting module: %v\n", err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("error reading response body: %v\n", err)
		}
		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			log.Fatalf("error reading zip: %v\n", err)
		}

		fileContents := make([]string, 0)
		for _, file := range zipReader.File {
			//fmt.Println(file.Name)
			if strings.HasSuffix(file.Name, ".go") && !strings.HasSuffix(file.Name, "_test.go") {
				stripped := strings.TrimPrefix(file.Name, fmt.Sprintf("%s@%s/", nName, dep.Version))
				//if !strings.Contains(stripped, "/") {
				if strings.Contains(stripped, "@") || strings.HasPrefix(stripped, "cmd") || strings.Contains(stripped, "example") {
					continue
				}
				fileContents = append(fileContents, stripped)
				//}
			}
		}
		resp.Body.Close()

		sum := sha256.Sum256(body)

		archiveRule := rule.NewRule("http_archive", fmt.Sprintf("%s@%s.mod", canonName, dep.Version))

		archiveRule.SetAttr("strip_prefix", fmt.Sprintf("%s@%s", nName, dep.Version))
		archiveRule.SetAttr("visibility", []string{"PUBLIC"})
		archiveRule.SetAttr("sha256", fmt.Sprintf("%x", sum))
		archiveRule.SetAttr("sub_targets", fileContents)
		archiveRule.SetAttr("urls", []string{fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.zip", nName, dep.Version)})

		depsOfCurrent := graph[fmt.Sprintf("%s@%s", dep.Path, dep.Version)]
		//goLibraryRule := rule.NewRule("go_library", fmt.Sprintf("%s@%s", canonName, dep.Version))
		goLibraryRule := rule.NewRule("go_library", canonName)
		deps := make([]string, len(depsOfCurrent))
		for key, value := range depsOfCurrent {
			split := strings.Split(value, "@")
			deps[key] = fmt.Sprintf("//third-party:%s", geomys.CanonicalizeModuleName(split[0]))
		}

		srcs := make([]string, 0)
		for _, file := range fileContents {
			srcs = append(srcs, fmt.Sprintf(":%s@%s.mod[%s]", canonName, dep.Version, file))
		}
		goLibraryRule.SetAttr("srcs", srcs)
		goLibraryRule.SetAttr("deps", deps)
		goLibraryRule.SetAttr("package_name", dep.Path)
		goLibraryRule.SetAttr("visibility", []string{"PUBLIC"})
		//glob := rule.GlobValue{
		//	Patterns: []string{fmt.Sprintf("$(location :%s@%s.mod)/*.go", canonName, dep.Version)},
		//	Excludes: []string{fmt.Sprintf("$(location :%s@%s.mod)/*_test.go", canonName, dep.Version)},
		//}
		rules = append(rules, archiveRule)
		rules = append(rules, goLibraryRule)
		rulesIndex++
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
