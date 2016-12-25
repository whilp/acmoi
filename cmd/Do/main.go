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

	name := os.Getenv("%")
	winid, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		return err
	}
	if *fOnce {
		return handle(name, winid)
	}

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
			handle(event.Name, event.ID)
		}
	}
}

func handle(name string, id int) error {
	win, err := acme.Open(id, nil)
	if err != nil {
		return err
	}
	defer win.CloseFiles()

	if err := format(name, win); err != nil {
		return err
	}

	if err := check(name, win); err != nil {
		return err
	}
	return nil
}

func check(name string, win *acme.Win) error {
	return nil
}

func format(name string, win *acme.Win) error {
	body, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	// TODO do this from the project root (ie, git toplevel)
	cmd := exec.Command("acme-format", name)
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
	// if _, err := win.Write("ctl", []byte("nomark\n")); err != nil {
	//	return err
	//}
	// defer func() { win.Write("ctl", []byte("mark\n")) }()

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

type project struct {
	root string
}

func NewProject(path string) *project {
	r, err := root(path)
	if err != nil {
		r = os.Getenv("HOME")
	}
	return &project{
		root: r,
	}
}

func root(path string) (string, error) {
	dir, _ := filepath.Split(path)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		out = out[:len(out)-3]
	}
	r := string(out)
	log.Printf("out %v", r)
	return string(out), err
}
