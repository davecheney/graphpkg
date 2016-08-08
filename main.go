// Graphpkg produces an svg graph of the dependency tree of a package
//
// Requires
// - dot (graphviz)
//
// Usage
//
//     graphpkg path/to/your/package
package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/browser"
)

var (
	pkgs     = make(map[string]map[string]struct{})
	seen     = make(map[string]struct{})
	matchvar = flag.String("match", ".*", "filter packages")
	maxLevel = flag.Uint("maxlevel", 0, "maximum package level to display")
	pkgmatch *regexp.Regexp
)

func truncate(p string) string {
	if *maxLevel == 0 {
		return p
	}

	slice := strings.Split(p, "/")
	lvl := int(*maxLevel)
	if len(slice) < lvl {
		lvl = len(slice)
	}

	return strings.Join(slice[:lvl], "/")
}

func truncateList(parent string, s []string) []string {
	if *maxLevel == 0 {
		return s
	}

	r := make([]string, 0, len(s))
	for _, p := range s {
		child := truncate(p)
		if child == parent {
			continue
		}
		r = append(r, child)
	}

	return r
}

func add(m map[string]struct{}, list ...string) map[string]struct{} {
	if m == nil {
		m = make(map[string]struct{})
	}
	for _, i := range list {
		m[i] = struct{}{}
	}
	return m
}

func findImport(p string) {
	t := truncate(p)

	if !pkgmatch.MatchString(p) {
		// doesn't match the filter, skip it
		return
	}
	if p == "C" {
		// C isn't really a package
		pkgs["C"] = nil
	}
	if _, ok := seen[p]; ok {
		// seen this package before, skip it
		return
	}
	pkg, err := build.Import(p, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	deps := filter(pkg.Imports)
	pkgs[t] = add(pkgs[t], truncateList(t, deps)...)

	for _, pkg := range deps {
		findImport(pkg)
	}
	seen[p] = struct{}{}
}

func filter(s []string) []string {
	var r []string
	for _, v := range s {
		if pkgmatch.MatchString(v) {
			r = append(r, v)
		}
	}
	return r
}

func allKeys() []string {
	keys := make(map[string]bool)
	for k, v := range pkgs {
		keys[k] = true
		for v, _ := range v {
			keys[v] = true
		}
	}
	v := make([]string, 0, len(keys))
	for k, _ := range keys {
		v = append(v, k)
	}
	return v
}

func keys() map[string]int {
	m := make(map[string]int)
	for i, k := range allKeys() {
		m[k] = i
	}
	return m
}

func init() {
	flag.Parse()
	pkgmatch = regexp.MustCompile(*matchvar)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	for _, pkg := range flag.Args() {
		findImport(pkg)
	}
	cmd := exec.Command("dot", "-Tsvg")
	in, err := cmd.StdinPipe()
	check(err)
	out, err := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	check(cmd.Start())

	fmt.Fprintf(in, "digraph {\n")
	keys := keys()
	for p, i := range keys {
		fmt.Fprintf(in, "\tN%d [label=%q,shape=box];\n", i, p)
	}
	for k, v := range pkgs {
		for p, _ := range v {
			fmt.Fprintf(in, "\tN%d -> N%d [weight=1];\n", keys[k], keys[p])
		}
	}
	fmt.Fprintf(in, "}\n")
	in.Close()

	ch := make(chan error)
	go func() {
		ch <- browser.OpenReader(out)

	}()
	check(cmd.Wait())
	if err := <-ch; err != nil {
		log.Fatalf("unable to open browser: %s", err)
	}
}
