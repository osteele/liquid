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

	"github.com/osteele/liquid"
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
		fmt.Fprint(stderr, err) // nolint: gas
		exit(1)
	}
}

func run(args []string) error {
	switch {
	case len(args) == 0:
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, stdin); err != nil {
			return err
		}
		render(buf.Bytes(), "")
	case args[0] == "-h" || args[0] == "--help":
		usage(false)
	case strings.HasPrefix(args[0], "-"):
		usage(true)
		exit(1)
	case len(args) == 1:
		s, err := ioutil.ReadFile(args[0])
		if err != nil {
			return err
		}
		render(s, args[0])
	default:
		usage(true)
	}
	return nil
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprint(stdout, err) // nolint: gas
		exit(1)
	}
}

func render(b []byte, filename string) {
	tpl, err := liquid.NewEngine().ParseTemplate(b)
	exitIfErr(err)
	out, err := tpl.Render(map[string]interface{}{})
	exitIfErr(err)
	stdout.Write(out) // nolint: gas, errcheck
}

func usage(error bool) {
	fmt.Printf("usage: %s [FILE]\n", os.Args[0])
	if error {
		exit(1)
	}
}
