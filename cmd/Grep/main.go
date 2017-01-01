/*

Grep searches for text within a project.

Grep chooses search text based on the selection within the current acme window. It then invokes acme-grep with the search text as its sole argument.

*/

package main // import "github.com/whilp/acmoi/cmd/Grep"

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
