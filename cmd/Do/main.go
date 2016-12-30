package main

// TODO run stuff async
// TODO ala davidrjenni/A, if invoked as Do, read selection and call out to acme-define (or inspect or something; for go, call guru).
// TODO Get strategy breaks undo across puts
// TODO gogetdoc
// TODO split into commands:
//  - Watch (runs Handle on put events)
//  - Grep (runs acme-grep in a project dir)
//  - Define (runs acme-define) describe?
//  - Test
//  - Build
// TODO shorten names relative to dir w/ filepath.Rel

import (
	"bytes"
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
		name := os.Getenv("%")
		winid, err := strconv.Atoi(os.Getenv("winid"))
		if err != nil {
			return err
		}
		h, err := handle(name, winid)
		if err != nil {
			return err
		}
		return h.define(name)
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
			h, err := handle(event.Name, event.ID)
			if err != nil {
				log.Print(err)
			}
			if err := h.Handle(); err != nil {
				log.Print(err)
			}
		}
	}
}

func handle(name string, id int) (*Handler, error) {
	var (
		err error
		h   *Handler
	)

	name, err = filepath.Abs(name)
	if err != nil {
		return h, err
	}

	win, err := acme.Open(id, nil)
	if err != nil {
		return h, err
	}
	defer win.CloseFiles()

	dir := toplevel(name) + "/"
	_, err = WindowByName(dir)
	if err != nil {
		return h, err
	}

	did, err := WindowID(dir)
	if err != nil {
		return h, err
	}
	e, err := NewErrors(did)
	if err != nil {
		return h, err
	}
	return NewHandler(name, id, e, dir, win), nil
}

// Handler receives acme log events and takes actions in response.
type Handler struct {
	name string
	id   int
	win  *acme.Win
	out  io.Writer
	dir  string
}

// NewHandler creates a new handler.
func NewHandler(name string, id int, out io.Writer, dir string, win *acme.Win) *Handler {
	return &Handler{
		name: name,
		id:   id,
		win:  win,
		out:  out,
		dir:  filepath.Clean(dir),
	}
}

// Handle runs format, check, build, and test.
func (h *Handler) Handle() error {
	if err := h.format(h.name); err != nil {
		return err
	}

	if err := h.check(h.name); err != nil {
		return err
	}

	if err := h.build(h.name); err != nil {
		return err
	}

	if err := h.test(h.name); err != nil {
		return err
	}
	return nil
}

func (h *Handler) run(cmd string, name string, args ...string) *exec.Cmd {
	// If we can't make name relative to our directory, roll with what we get.
	n, err := filepath.Rel(h.dir, filepath.Clean(name))
	if err == nil {
		name = n
	}

	args = append([]string{name}, args...)
	c := exec.Command(cmd, args...)
	c.Dir = h.dir
	c.Stderr = h.out
	c.Stdout = h.out
	return c
}

func (h *Handler) define(name string) error {
	pos, _, err := selection(h.win)
	if err != nil {
		return err
	}

	a := NewArchive()

	windows, err := acme.Windows()
	if err != nil {
		return err
	}

	for _, wi := range windows {
		w, err := acme.Open(wi.ID, nil)
		if err != nil {
			return err
		}
		c, err := NewCtl(w)
		if err != nil {
			return err
		}
		if c.IsDirectory {
			continue
		}

		b, err := w.ReadAll("body")
		if err != nil {
			return err
		}
		if err := a.Write(c.Name, b); err != nil {
			return err
		}
	}

	cmd := h.run("acme-define", name, strconv.Itoa(pos))
	cmd.Stdin = a.Buffer()
	return cmd.Run()
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
	// TODO figure out why format isn't passing short names
	cmd := h.run("acme-format", name)

	if err := cmd.Run(); err != nil {
		return err
	}

	q0, q1, err := selection(h.win)
	if err != nil {
		return err
	}

	if err := h.win.Ctl("get\n"); err != nil {
		return err
	}

	if err := show(h.win, q0, q1); err != nil {
		return err
	}

	if err := h.win.Ctl("clean\n"); err != nil {
		return err
	}

	return nil
}

func selection(win *acme.Win) (int, int, error) {
	// First addr read is always 0.
	if _, _, err := win.ReadAddr(); err != nil {
		return 0, 0, err
	}

	if err := win.Ctl("addr=dot\n"); err != nil {
		return 0, 0, err
	}

	q0, q1, err := win.ReadAddr()
	if err != nil {
		return 0, 0, err
	}

	return q0, q1, nil
}

func show(win *acme.Win, q0, q1 int) error {
	if err := win.Addr("#%d,#%d", q0, q1); err != nil {
		return err
	}

	if err := win.Ctl("dot=addr\nshow\n"); err != nil {
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
	return filepath.Clean(strings.Trim(string(out), "\n"))
}

type Archive struct {
	buf *bytes.Buffer
}

func NewArchive() *Archive {
	buf := new(bytes.Buffer)
	return &Archive{buf: buf}
}

func (a *Archive) Write(name string, body []byte) error {
	header := fmt.Sprintf("%s\n%d\n", name, len(body))
	if _, err := a.buf.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := a.buf.Write(body); err != nil {
		return err
	}
	return nil
}

func (a *Archive) Buffer() *bytes.Reader {
	return bytes.NewReader(a.buf.Bytes())
}

// TODO contribute this to acme
var fsys *client.Fsys
var fsysErr error
var fsysOnce sync.Once

func mountAcme() {
	fsys, fsysErr = client.MountService("acme")
}

// Cons is a special file used by acme to log errors.
type Cons struct {
	f *client.Fid
}

// NewCons returns a new cons file.
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

// Close closes the cons file.
func (c *Cons) Close() error {
	return c.f.Close()
}

// Write writes a message to the cons file.
func (c *Cons) Write(b []byte) (int, error) {
	return c.f.Write(b)
}

// Errors is a special file associated with each acme window for logging errors.
type Errors struct {
	f *client.Fid
}

// NewErrors returns an errors file.
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

// Close closes the errors file.
func (e *Errors) Close() error {
	return e.f.Close()
}

// Write writes a message to the errors file.
func (e *Errors) Write(b []byte) (int, error) {
	return e.f.Write(b)
}

// WindowByName finds an open window for the given file name, or creates a new one if no such window is found.
func WindowByName(name string) (*acme.Win, error) {
	windows, err := acme.Windows()
	if err != nil {
		return nil, err
	}
	a, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	for _, w := range windows {
		b, err := filepath.Abs(w.Name)
		_ = err // ignore err
		if a == b {
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

// WindowID returns the ID for a window that holds the named file.
func WindowID(name string) (int, error) {
	windows, err := acme.Windows()
	if err != nil {
		return 0, err
	}
	a, err := filepath.Abs(name)
	if err != nil {
		return 0, err
	}
	for _, w := range windows {
		b, err := filepath.Abs(w.Name)
		_ = err // ignore err here
		if a == b {
			return w.ID, nil
		}
	}
	return 0, fmt.Errorf("could not find id for %s", name)
}

// WindowName returns the name of a file with the given ID.
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

type Ctl struct {
	ID          int
	TagSize     int
	BodySize    int
	IsDirectory bool
	IsModified  bool
	Tag         string
	Name        string
}

func NewCtl(win *acme.Win) (*Ctl, error) {
	c := &Ctl{}
	ctl, err := win.ReadAll("ctl")
	if err != nil {
		return c, err
	}
	fields := strings.Fields(string(ctl))

	if c.ID, err = strconv.Atoi(fields[0]); err != nil {
		return c, err
	}
	if c.TagSize, err = strconv.Atoi(fields[1]); err != nil {
		return c, err
	}
	if c.BodySize, err = strconv.Atoi(fields[2]); err != nil {
		return c, err
	}
	c.IsDirectory = fields[3] == "1"
	c.IsModified = fields[4] == "1"

	tag, err := win.ReadAll("tag")
	if err != nil {
		return c, err
	}
	c.Tag = string(tag)
	c.Name = strings.Fields(c.Tag)[0]
	return c, nil
}
