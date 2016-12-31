package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"9fans.net/go/acme"
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
	name := flag.Arg(0)
	clean, err := filepath.EvalSymlinks(name)
	if err == nil {
		name = clean
	}
	abs, err := filepath.Abs(name)
	if err == nil {
		name = abs
	}

	cons, err := acmoi.NewCons()
	if err != nil {
		return err
	}
	defer cons.Close()
	log.SetOutput(cons)

	win, err := acme.New()
	if err != nil {
		return err
	}

	if err := win.Name(name); err != nil {
		return err
	}

	w, err := acmoi.NewWindow(win)
	if err != nil {
		return err
	}

	parent, err := w.Parent()
	if err != nil {
		return err
	}
	w.Errors = parent.Errors

	return edit(w)
}

func edit(win *acmoi.Window) error {
	events, err := acme.Log()
	if err != nil {
		return err
	}
	defer events.Close()
	if err := win.Win.Ctl("get\n"); err != nil {
		return err
	}

	for {
		event, err := events.Read()
		if err != nil {
			return err
		}
		log.Printf("event %v", event)

		if event.Op == "del" && event.Name == win.Ctl.Name() {
			return nil
		}
	}
	return fmt.Errorf("missed del for %s", win.Ctl.Name())
}
