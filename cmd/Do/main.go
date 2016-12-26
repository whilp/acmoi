package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"9fans.net/go/acme"
	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
)

var (
	fDaemon = flag.Bool("daemon", false, "run in the background")
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	flag.Parse()

	cons, err := NewCons()
	if err != nil {
		return err
	}
	defer cons.Close()
	log.SetOutput(cons)

	if !*fDaemon {
		return once()
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
			err := handle(event.Name, event.ID)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func once() error {
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

	if _, err := win.Write("ctl", []byte("put\n")); err != nil {
		return err
	}

	return handle(name, winid)
}

func handle(name string, id int) error {
	var err error

	name, err = filepath.Abs(name)
	if err != nil {
		return err
	}

	win, err := acme.Open(id, nil)
	if err != nil {
		return err
	}
	defer win.CloseFiles()

	dir := toplevel(name) + "/"
	_, err = WindowByName(dir)
	if err != nil {
		return err
	}

	did, err := WindowID(dir)
	if err != nil {
		return err
	}
	e, err := NewErrors(did)
	if err != nil {
		log.Print(err)
		return err
	}
	handler := NewHandler(e, dir, win)
	return handler.Handle(name, id)
}

type Handler struct {
	win *acme.Win
	out io.Writer
	dir string
}

func NewHandler(out io.Writer, dir string, win *acme.Win) *Handler {
	return &Handler{
		win: win,
		out: out,
		dir: dir,
	}
}

func (h *Handler) Handle(name string, id int) error {
	if err := h.format(name); err != nil {
		return err
	}

	if err := h.build(name); err != nil {
		return err
	}

	if err := h.test(name); err != nil {
		log.Printf("an err %s", err)
		return err
	}
	return nil
}

func (h *Handler) run(cmd string, args ...string) *exec.Cmd {
	c := exec.Command(cmd, args...)
	c.Dir = h.dir
	c.Stderr = h.out
	c.Stdout = h.out
	return c
}

func (h *Handler) check(name string) error {
	return h.run("acme-check", name).Run()
}

func (h *Handler) build(name string) error {
	return h.run("acme-build", name).Run()
}

func (h *Handler) test(name string) error {
	return h.run("acme-test", name).Run()
}

func (h *Handler) format(name string) error {
	cmd := h.run("acme-format", name)

	if err := cmd.Run(); err != nil {
		return err
	}

	// First addr read is always 0.
	if _, _, err := h.win.ReadAddr(); err != nil {
		return err
	}

	if _, err := h.win.Write("ctl", []byte("addr=dot\n")); err != nil {
		return err
	}

	q0, q1, err := h.win.ReadAddr()
	if err != nil {
		return err
	}

	if _, err := h.win.Write("ctl", []byte("get\n")); err != nil {
		return err
	}

	if err := show(h.win, q0, q1); err != nil {
		return err
	}

	if _, err := h.win.Write("ctl", []byte("clean\n")); err != nil {
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

func toplevel(path string) string {
	dir := filepath.Dir(path)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return dir
	}
	return strings.Trim(string(out), "\n")
}

// TODO contribute this to acme
var fsys *client.Fsys
var fsysErr error
var fsysOnce sync.Once

func mountAcme() {
	fsys, fsysErr = client.MountService("acme")
}

type Cons struct {
	f *client.Fid
}

func NewCons() (*Cons, error) {
	fsysOnce.Do(mountAcme)
	if fsysErr != nil {
		return nil, fsysErr
	}
	f, err := fsys.Open("cons", plan9.ORDWR)
	if err != nil {
		return nil, err
	}
	return &Cons{f: f}, nil
}

func (c *Cons) Close() error {
	return c.f.Close()
}

func (c *Cons) Write(b []byte) (int, error) {
	return c.f.Write(b)
}

type Errors struct {
	f *client.Fid
}

func NewErrors(id int) (*Errors, error) {
	fsysOnce.Do(mountAcme)
	if fsysErr != nil {
		return nil, fsysErr
	}

	name := fmt.Sprintf("%d/errors", id)
	f, err := fsys.Open(name, plan9.OWRITE)
	if err != nil {
		return nil, err
	}

	return &Errors{f: f}, nil
}

func (e *Errors) Close() error {
	return e.f.Close()
}

func (e *Errors) Write(b []byte) (int, error) {
	return e.f.Write(b)
}

func WindowByName(name string) (*acme.Win, error) {
	windows, err := acme.Windows()
	if err != nil {
		return nil, err
	}
	for _, w := range windows {
		if w.Name == name {
			return acme.Open(w.ID, nil)
		}
	}

	win, err := acme.New()
	if err != nil {
		return nil, err
	}
	if err := win.Name(name); err != nil {
		return nil, err
	}
	return win, nil
}

func WindowID(name string) (int, error) {
	windows, err := acme.Windows()
	if err != nil {
		return 0, err
	}
	for _, w := range windows {
		if w.Name == name {
			return w.ID, nil
		}
	}
	return 0, fmt.Errorf("could not find id for %s", name)
}

func WindowName(id int) (string, error) {
	windows, err := acme.Windows()
	if err != nil {
		return "", err
	}
	for _, w := range windows {
		if w.ID == id {
			return w.Name, nil
		}
	}
	return "", fmt.Errorf("could not find name for %d", id)
}
