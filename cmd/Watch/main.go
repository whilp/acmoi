// Copyright 2016 by whilp.

/*

Watch calls various helpers on put events emitted by acme.

When it sees such an event, it attempts to invoke each of the following external executables until it encounters an error:

- `acme-format`
- `acme-check`
- `acme-build`
- `acme-test`

First, Watch determines the root of the project containing the file that has just been written by invoking `acme-root` with no arguments. Then, it invokes each command in sequence until it encounters an error, passing the path to the newly written file relative to the project root. The stdout and stderr of the commands are routed to the project root's +Errors window. `acme-format` is expected to rewrite the file on disk; if it succeeds, the window containing the file will be refreshed. Any command may return a non-zero status to stop further execution.

*/

package main

// TODO run stuff async
// TODO Get strategy breaks undo across puts

import (
	"flag"
	"log"

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
	cons, err := acmoi.NewCons()
	if err != nil {
		return err
	}
	defer cons.Close()
	log.SetOutput(cons)

	events, err := acme.Log()
	if err != nil {
		return err
	}
	defer events.Close()

	for {
		event, err := events.Read()
		if err != nil {
			return err
		}
		if event.Op == "put" && event.Name != "" {
			win, err := acmoi.NewWindowFromID(event.ID)
			_ = err
			// if err != nil {
			//	log.Print(err)
			// }
			err = handle(win)
			_ = err
			// if err != nil {
			//	log.Print(err)
			//}
		}
	}
}

func handle(win *acmoi.Window) error {
	parent, err := win.Parent()
	if err != nil {
		return err
	}
	win.Errors = parent.Errors

	if err := format(win); err != nil {
		return err
	}

	if err := check(win); err != nil {
		return err
	}

	if err := build(win); err != nil {
		return err
	}

	if err := test(win); err != nil {
		return err
	}
	return nil
}

func check(win *acmoi.Window) error {
	return win.Do("acme-check", win.Rel()).Run()
}

func build(win *acmoi.Window) error {
	return win.Do("acme-build", win.Rel()).Run()
}

func test(win *acmoi.Window) error {
	return win.Do("acme-test", win.Rel()).Run()
}

func format(win *acmoi.Window) error {
	cmd := win.Do("acme-format", win.Rel())
	if err := cmd.Run(); err != nil {
		return err
	}

	q0, q1, err := win.Selection()
	if err != nil {
		return err
	}

	if err := win.Win.Ctl("get\n"); err != nil {
		return err
	}

	if err := show(win, q0, q1); err != nil {
		return err
	}

	if err := win.Win.Ctl("clean\n"); err != nil {
		return err
	}

	return nil
}

func show(win *acmoi.Window, q0, q1 int) error {
	if err := win.Addr("#%d,#%d", q0, q1); err != nil {
		return err
	}

	if err := win.Win.Ctl("dot=addr\nshow\n"); err != nil {
		return err
	}
	return nil
}
