package main

import (
	"flag"
	"log"

	"github.com/whilp/acmoi"
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
	return grep(w)
}

func grep(win *acmoi.Window) error {
	from, to, err := win.Selection()
	if err != nil {
		return err
	}

	body, err := win.ReadAll("body")
	if err != nil {
		return err
	}
	pattern := string(body[from:to])

	cmd := win.Do("acme-grep", pattern)
	return cmd.Run()
}
