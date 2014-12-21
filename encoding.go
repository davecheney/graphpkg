package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func allKeys(pkgs *map[string][]string) []string {
	keys := make(map[string]bool)
	for k, v := range *pkgs {
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

func keys(pkgs *map[string][]string) map[string]int {
	m := make(map[string]int)
	for i, k := range allKeys(pkgs) {
		m[k] = i
	}
	return m
}

// writeDotRaw outputs the import graph as dot
func writeDotRaw(w io.Writer, pkgs *map[string][]string) error {
	fmt.Fprint(w, "digraph {\n")
	keys := keys(pkgs)
	for p, i := range keys {
		fmt.Fprintf(w, "\tN%d [label=%q,shape=box];\n", i, p)
	}
	for k, v := range *pkgs {
		for _, p := range v {
			fmt.Fprintf(w, "\tN%d -> N%d [weight=1];\n", keys[k], keys[p])
		}
	}
	fmt.Fprintf(w, "}\n")
	return nil
}

// writeDotOutput uses graphviz's dot util to convert dot fmt into something else
func writeDotOutput(out io.Writer, fmt string, pkgs *map[string][]string) error {
	cmd := exec.Command("dot", "-T"+fmt)
	cmd.Stdout = out // write dot's output straight to ours.
	cmd.Stderr = os.Stderr

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// write dot fmt into dot's input
	if err := writeDotRaw(in, pkgs); err != nil {
		return err
	}
	in.Close()
	return cmd.Wait()
}

// writeSVG uses graphviz's dot util to convert dot fmt into svg
func writeSVG(out io.Writer, pkgs *map[string][]string) error {
	return writeDotOutput(out, "svg", pkgs)
}

type d3pkg struct {
	Name    string   `json:"name,omitempty"`
	Size    int      `json:"size,omitempty"`
	Imports []string `json:"imports,omitempty"`
}

// writeD3JSON outputs the import graph as mbostock's json imports thing
func writeD3JSON(w io.Writer, pkgs *map[string][]string) error {
	d3pkgs := pkgsToD3Pkgs(pkgs)
	enc := json.NewEncoder(w)

	for _, p := range *d3pkgs {
		if err := enc.Encode(p); err != nil {
			return err
		}
	}
	return nil
}

func pkgsToD3Pkgs(pkgs *map[string][]string) *map[string]d3pkg {
	d3pkgs := make(map[string]d3pkg)
	for p, imports := range *pkgs {
		d3pkgs[p] = d3pkg{
			Name:    p,
			Size:    1000, // change this once we know wtf it is.
			Imports: imports,
		}
	}
	return &d3pkgs
}
