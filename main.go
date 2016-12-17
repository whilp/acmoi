package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"

	"9fans.net/go/acme"
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

	win, err := acme.Open(winid, nil)
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

	if err := win.Addr("0,$"); err != nil {
		return err
	}

	if !bytes.Equal(body, result) {
		_, err := win.Write("data", result)
		if err != nil {
			return err
		}

		if _, err := win.Write("ctl", []byte("put\n")); err != nil {
			return err
		}

		if err := win.Addr("#%d,#%d", q0, q1); err != nil {
			return err
		}

		if _, err := win.Write("ctl", []byte("dot=addr\nshow\n")); err != nil {
			return err
		}
	}

	return nil
}
