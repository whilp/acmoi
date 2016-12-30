package main

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

func rel(win *acmoi.Window) string {
	return win.FromRoot(win.Ctl.Name())
}
