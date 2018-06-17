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
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/browser"
)

var (
	pkgs     = make(map[string][]string)
	matchvar = flag.String("match", ".*", "filter packages")
	stdout   = flag.Bool("stdout", false, "print to standard output instead of browser")
	pkgmatch *regexp.Regexp
)

func findImport(p string, rootp string) {
	if !pkgmatch.MatchString(p) {
		// doesn't match the filter, skip it
		return
	}
	if p == "C" {
		// C isn't really a package
		pkgs["C"] = nil
	}
	if _, ok := pkgs[p]; ok {
		// seen this package before, skip it
		return
	}
	if strings.HasPrefix(p, "golang_org") {
		p = path.Join("vendor", p)
	}

	pkg, err := build.Import(p, rootp, 0)
	if err != nil {
		log.Fatal(err)
	}
	pkgs[p] = filter(pkg.Imports)
	for _, pkg := range pkgs[p] {
		findImport(pkg, rootp)
	}
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
		for _, v := range v {
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
	var rootPkg * build.Package
	var e error
	for _, pkg := range flag.Args() {
		rootPkg, e = build.Import(pkg, "", 0)
		if e != nil {
			log.Fatal(e)
		}
		findImport(pkg, rootPkg.Dir)
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
		for _, p := range v {
			fmt.Fprintf(in, "\tN%d -> N%d [weight=1];\n", keys[k], keys[p])
		}
	}
	fmt.Fprintf(in, "}\n")
	in.Close()

	if *stdout {
		// print to standard output
		io.Copy(os.Stdout, out)

	} else {
		// pipe output to browser
		ch := make(chan error)
		go func() {
			ch <- browser.OpenReader(out)

		}()
		check(cmd.Wait())
		if err := <-ch; err != nil {
			log.Fatalf("unable to open browser: %s", err)
		}
	}
}
