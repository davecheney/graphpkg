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
	"regexp"

	"github.com/pkg/browser"
)

var (
	pkgs       = make(map[string][]string)
	matchvar   = flag.String("match", ".*", "filter packages")
	browservar = flag.Bool("browser", false, "open a browser with the output")
	formatvar  = flag.String("format", "svg", "format: {svg, dot, d3json}")
	pkgmatch   *regexp.Regexp
)

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
	pkg, err := build.Import(p, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	pkgs[p] = filter(pkg.Imports)
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

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <package name>\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	pkgmatch = regexp.MustCompile(*matchvar)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, pkg := range args {
		findImport(pkg)
	}

	r, w, err := os.Pipe()
	check(err)

	// run the transform
	go xform(w)
	output(r)
}

func xform(w io.WriteCloser) {
	var err error
	switch *formatvar {
	case "d3json":
		err = writeD3JSON(w, &pkgs)
	case "svg":
		err = writeSVG(w, &pkgs)
	case "dot":
		err = writeDotRaw(w, &pkgs)
	default:
		err = fmt.Errorf("error: unknown format %s", *formatvar)
	}
	w.Close()
	check(err)
}

func output(r io.Reader) {
	var err error
	switch {
	case *browservar:
		fmt.Println("opening in your browser...")
		err = browser.OpenReader(r)
	default:
		_, err = io.Copy(os.Stdout, r)
	}
	check(err)
}
