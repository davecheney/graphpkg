package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var (
	pkgs = make(map[string][]string)
)

func findImport(p string) {
	if p == "C" {
		pkgs["C"] = nil
	}
	if _, ok := pkgs[p]; ok {
		return
	}
	pkg, err := build.Import(p, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	pkgs[p] = pkg.Imports
	for _, pkg := range pkg.Imports {
		findImport(pkg)
	}
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

func main() {
	findImport(os.Args[1])
	cmd := exec.Command("dot", "-Tsvg")
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	out, err := ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdout = out
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(in, "digraph %q {\n", os.Args[1])
	keys := keys()
	for p, i := range keys {
		fmt.Fprintf(in, "\tN%d [label=%q,shape=box,box];\n", i, p)
	}
	for k, v := range pkgs {
		for _, p := range v {
			fmt.Fprintf(in, "\tN%d -> N%d [weight=1];\n", keys[k], keys[p])
		}
	}
	fmt.Fprintf(in, "}\n")
	in.Close()
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	out.Close()
	os.Rename(out.Name(), out.Name()+".svg")
	fmt.Println(out.Name() + ".svg")
}
