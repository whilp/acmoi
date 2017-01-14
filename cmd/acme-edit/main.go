/*

acme-edit creates a new window for a file to be edited.

While the window is open, acme-edit waits, polling for a change in status. When the window is deleted, acme-edit exits. This makes acme-edit suitable for use as a git editor.

*/
package main // import "github.com/whilp/acmoi/cmd/acme-edit"

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

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

	if err := touch(name); err != nil {
		return err
	}

	win, err := acmoi.NewWindowFromName(name, true)
	if err != nil {
		return err
	}
	if err := win.Win.Ctl("get\n"); err != nil {
		return err
	}

	return wait(win.Ctl.ID())
}

func touch(name string) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte{})
	return err
}

func wait(id int) error {
	for {
		// Use acme.Open directly to reduce the likelihood of
		// panics. acmoi.Window helpfully creates the ctl and
		// errors structs, which requires a few extra reads and
		// increases the chance that such a read happens on an
		// invalid ID.
		_, err := acme.Open(id, nil)
		if err != nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func waitevent(name string) error {
	// TODO this should work, but acme seems to omit del events
	// for the window we create.
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
		if event.Op == "focus" {
			continue
		}
		log.Printf("event %v", event)

		if event.Op == "del" && event.Name == name {
			return nil
		}
	}
}
