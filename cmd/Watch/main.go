// Watch calls various helpers on put events emitted by acme.
//
// When it sees such an event, it attempts to invoke each of the following external executables until it encounters an error:
//
// 	acme-format
// 	acme-check
// 	acme-build
// 	acme-test
//
// First, Watch determines the root of the project containing the file that has just been written by invoking acme-root with no arguments. Then, it invokes each command in sequence until it encounters an error, passing the path to the newly written file relative to the project root. The stdout and stderr of the commands are routed to the project root's +Errors window. `acme-format` is expected to rewrite the file on disk; if it succeeds, the window containing the file will be refreshed. Any command may return a non-zero status to stop further execution.

package main // import "github.com/whilp/acmoi/cmd/Watch"

import (
	"bytes"
	"flag"
	"io/ioutil"
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
			if err != nil {
				log.Print(err)
			}
			go handle(win)
		}
	}
}

func handle(win *acmoi.Window) error {
	defer win.CloseFiles()
	if ignore(win) {
		return nil
	}
	parent, err := win.Parent()
	if err != nil {
		return err
	}
	win.Errors = parent.Errors
	defer parent.CloseFiles()

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
	before, err := ioutil.ReadFile(win.Ctl.Name())
	if err != nil {
		return err
	}

	cmd := win.Do("acme-format", win.Rel())
	if err := cmd.Run(); err != nil {
		return err
	}

	after, err := ioutil.ReadFile(win.Ctl.Name())
	if err != nil {
		return err
	}

	if bytes.Equal(before, after) {
		return nil
	}

	q0, q1, err := win.Selection()
	if err != nil {
		return err
	}

	w := &winderr{Window: win}
	w.ctl("get\n")
	w.addr("#%d,#%d", q0, q1)
	w.ctl("dot=addr\nshow\n")
	w.ctl("clean\n")

	return w.err
}

func ignore(win *acmoi.Window) bool {
	tag, err := win.ReadAll("tag")
	if err != nil {
		return true
	}
	if bytes.Contains(tag, []byte("!Watch")) {
		return true
	}
	return false
}

type winderr struct {
	*acmoi.Window
	err error
}

func (w *winderr) addr(format string, args ...interface{}) {
	if w.err != nil {
		return
	}
	w.err = w.Addr(format, args...)
}

func (w *winderr) ctl(format string, args ...interface{}) {
	if w.err != nil {
		return
	}
	w.err = w.Win.Ctl(format, args...)
}

func (w *winderr) write(file string, b []byte) {
	if w.err != nil {
		return
	}
	_, w.err = w.Write(file, b)
}
