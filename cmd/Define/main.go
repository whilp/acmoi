/*

Define prints basic information about currently selected syntax.

Define determines its input based on the selection in the acme window in which it is invoked. It then calls the acme-define executable, passing an archive formatted for the go guru tool on stdin. The archive contains all open acme buffers, with each entry consisting of the file name, a newline, the file size, another newline, and the contents of the file.

*/
package main // import "github.com/whilp/acmoi/cmd/Define"

import (
	"flag"
	"log"
	"strconv"

	"github.com/whilp/acmoi"

	"9fans.net/go/acme"
)

func main() {
	flag.Parse()
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cons, err := acmoi.NewCons()
	if err != nil {
		return err
	}
	defer cons.Close()
	log.SetOutput(cons)

	w, err := acmoi.NewWindowFromEnviron()
	if err != nil {
		return err
	}
	parent, err := w.Parent()
	if err != nil {
		return err
	}
	w.Errors = parent.Errors
	return define(w)
}

func define(win *acmoi.Window) error {
	pos, _, err := win.Selection()
	if err != nil {
		return err
	}

	a := acmoi.NewArchive()

	windows, err := acme.Windows()
	if err != nil {
		return err
	}

	for _, wi := range windows {
		w, err := acmoi.NewWindowFromID(wi.ID)
		if w.IsDirectory() {
			continue
		}

		b, err := w.ReadAll("body")
		if err != nil {
			return err
		}
		if err := a.Write(w.Ctl.Name(), b); err != nil {
			return err
		}
	}

	cmd := win.Do("acme-define", win.Rel(), strconv.Itoa(pos))
	cmd.Stdin = a.Buffer()
	return cmd.Run()
}
