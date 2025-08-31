// Package main defines a command-line interface to the Liquid engine.
//
// This command intended for testing and bug reports.
//
// Examples:
//
//	echo '{{ "Hello " | append: "World" }}' | liquid
//	liquid source.tpl
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/osteele/liquid"
)

// for testing
var (
	stderr     io.Writer       = os.Stderr
	stdout     io.Writer       = os.Stdout
	stdin      io.Reader       = os.Stdin
	exit       func(int)       = os.Exit
	env        func() []string = os.Environ
	bindings   map[string]any  = map[string]any{}
	strictVars bool
)

func main() {
	var err error

	cmdLine := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	cmdLine.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s [OPTIONS] [FILE]\n", cmdLine.Name()) //nolint:errcheck
		fmt.Fprint(stderr, "\nOPTIONS\n")                                   //nolint:errcheck
		cmdLine.PrintDefaults()
	}

	var bindEnvs bool
	cmdLine.BoolVar(&bindEnvs, "env", false, "bind environment variables")
	cmdLine.BoolVar(&strictVars, "strict", false, "enable strict variable mode in templates")

	err = cmdLine.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			exit(0)
			return
		}
		fmt.Fprintln(stderr, err) //nolint:errcheck
		exit(1)
		return
	}

	if bindEnvs {
		for _, e := range env() {
			pair := strings.SplitN(e, "=", 2)
			bindings[pair[0]] = pair[1]
		}
	}

	args := cmdLine.Args()
	switch len(args) {
	case 0:
		// use stdin
	case 1:
		stdin, err = os.Open(args[0])
	default:
		err = errors.New("too many arguments")
	}

	if err == nil {
		err = render()
	}

	if err != nil {
		fmt.Fprintln(stderr, err) //nolint:errcheck
		exit(1)
	}
}

func render() error {
	buf, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}

	e := liquid.NewEngine()
	if strictVars {
		e.StrictVariables()
	}
	tpl, err := e.ParseTemplate(buf)
	if err != nil {
		return err
	}
	out, err := tpl.Render(bindings)
	if err != nil {
		return err
	}
	_, err = stdout.Write(out)
	return err
}
