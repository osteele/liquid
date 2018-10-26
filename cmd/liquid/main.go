// Package main defines a command-line interface to the Liquid engine.
//
// This command intended for testing and bug reports.
//
// Examples:
//
// 	echo '{{ "Hello " | append: "World" }}' | liquid
// 	liquid source.tpl
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/urbn8/liquid"
)

// for testing
var (
	stderr           = os.Stderr
	stdout io.Writer = os.Stdout
	stdin  io.Reader = os.Stdin
	exit             = os.Exit
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(stderr, err) // nolint: gas
		os.Exit(1)
	}
}

func run(args []string) error {
	switch {
	case len(args) == 0:
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, stdin); err != nil {
			return err
		}
		return render(buf.Bytes(), "")
	case args[0] == "-h" || args[0] == "--help":
		usage()
	case strings.HasPrefix(args[0], "-"):
		// undefined flag
		usage()
		exit(1)
	case len(args) == 1:
		s, err := ioutil.ReadFile(args[0])
		if err != nil {
			return err
		}
		return render(s, args[0])
	default:
		usage()
		exit(1)
	}
	return nil
}

func render(b []byte, filename string) (err error) {
	tpl, err := liquid.NewEngine().ParseTemplate(b)
	if err != nil {
		return err
	}
	out, err := tpl.Render(map[string]interface{}{})
	if err != nil {
		return err
	}
	_, err = stdout.Write(out)
	return err
}

func usage() {
	fmt.Fprintf(stdout, "usage: %s [FILE]\n", os.Args[0]) // nolint: gas
}
