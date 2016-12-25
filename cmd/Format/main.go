package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"9fans.net/go/acme"
)

var (
	fOnce = flag.Bool("once", false, "run once")
)

func main() {
	err := inner()
	if err != nil {
		log.Fatal(err)
	}
}

func inner() error {
	flag.Parse()
	args := flag.Args()

	name := os.Getenv("%")
	winid, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		return err
	}
	if *fOnce {
		return format(name, winid, args)
	} else {
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
				autoFormat(event.Name, event.ID)
			}
		}
	}
}

func autoFormat(name string, id int) error {
	ext := filepath.Ext(name)
	formatters := []struct {
		ext  string
		args []string
	}{
		{".go", []string{"goreturns", "-i"}},
	}

	for _, formatter := range formatters {
		if formatter.ext == ext {
			return format(name, id, formatter.args)
		}
	}
	return nil
}

func format(name string, id int, args []string) error {
	win, err := acme.Open(id, nil)
	if err != nil {
		return err
	}
	defer win.CloseFiles()

	body, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	args = append(args, name)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr

	result, err := cmd.Output()
	if err != nil {
		return err
	}

	// First addr read is always 0.
	if _, _, err := win.ReadAddr(); err != nil {
		return err
	}

	if _, err := win.Write("ctl", []byte("addr=dot\n")); err != nil {
		return err
	}

	q0, q1, err := win.ReadAddr()
	if err != nil {
		return err
	}

	if bytes.Equal(body, result) {
		return nil
	}
	if err := rewrite(win, result); err != nil {
		return err
	}
	if err := show(win, q0, q1); err != nil {
		return err
	}
	if err := put(win); err != nil {
		return err
	}

	return nil
}

func rewrite(win *acme.Win, body []byte) error {
	if _, err := win.Write("ctl", []byte("nomark\n")); err != nil {
		return err
	}
	defer func() { win.Write("ctl", []byte("mark\n")) }()

	if err := win.Addr("0,$"); err != nil {
		return err
	}

	_, err := win.Write("data", body)
	if err != nil {
		return err

	}

	return nil
}

func show(win *acme.Win, q0, q1 int) error {
	if err := win.Addr("#%d,#%d", q0, q1); err != nil {
		return err
	}

	if _, err := win.Write("ctl", []byte("dot=addr\nshow\n")); err != nil {
		return err
	}
	return nil
}

func put(win *acme.Win) error {
	_, err := win.Write("ctl", []byte("put\n"))
	return err
}
