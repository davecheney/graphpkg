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
	tags     stringSlice
	ctx      build.Context
)

type stringSlice []string

func (ss *stringSlice) String() string {
	return strings.Join(*ss, ",")
}

func (ss *stringSlice) Set(s string) error {
	*ss = append(*ss, strings.Split(s, ",")...)
	return nil
}

func findImport(p string) {
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

	wd, _ := os.Getwd()
	pkg, err := ctx.Import(p, wd, 0)
	if err != nil {
		log.Fatal(err)
	}
	allImports := pkg.Imports
	allImports = append(allImports, pkg.TestImports...)
	allImports = append(allImports, pkg.XTestImports...)
	pkgs[p] = filter(allImports)
	for _, pkg := range pkgs[p] {
		findImport(pkg)
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
	for k := range keys {
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
	flag.Var(&tags, "tags", "a list of build tags to consider satisfied during the build")
	flag.Parse()
	pkgmatch = regexp.MustCompile(*matchvar)
	ctx = build.Default
	ctx.BuildTags = tags
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
	out, _ := cmd.StdoutPipe()
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
