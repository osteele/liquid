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

func main() {
	args := os.Args[1:]
	switch {
	case len(args) == 0:
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, os.Stdin)
		exitIfErr(err)
		render(buf.Bytes(), "")
	case args[0] == "-h" || args[0] == "--help":
		usage(false)
	case strings.HasPrefix(args[0], "-"):
		usage(true)
		os.Exit(1)
	case len(args) == 1:
		s, err := ioutil.ReadFile(args[0])
		exitIfErr(err)
		render(s, args[0])
	default:
		usage(true)
	}
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprint(os.Stdout, err) // nolint: gas
		os.Exit(1)
	}
}

func render(b []byte, filename string) {
	tpl, err := liquid.NewEngine().ParseTemplate(b)
	exitIfErr(err)
	tpl.SetSourcePath(filename)
	out, err := tpl.Render(map[string]interface{}{})
	exitIfErr(err)
	os.Stdout.Write(out) // nolint: gas, errcheck
}

func usage(error bool) {
	fmt.Printf("usage: %s [FILE]\n", os.Args[0])
	if error {
		os.Exit(1)
	}
}
