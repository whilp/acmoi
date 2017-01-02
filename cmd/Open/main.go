/*

Open creates a new window for a file to be edited.

While the window is open, Open waits, polling for a change in status. When the window is deleted, Open exits. This makes Open suitable for use as a git editor.

*/
package main // import "github.com/whilp/acmoi/cmd/Open"

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
		_, err := acmoi.NewWindowFromID(id)
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
