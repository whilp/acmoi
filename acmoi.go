package acmoi

import (
	"fmt"
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

// TODO contribute this to acme
var fsys *client.Fsys
var fsysErr error
var fsysOnce sync.Once

func mountAcme() {
	fsys, fsysErr = client.MountService("acme")
}

type Window struct {
	*acme.Win
	*Ctl
	Errors *Errors
	Root   string
}

func NewWindow(win *acme.Win) (*Window, error) {
	w := &Window{}
	var err error
	if w.Ctl, err = NewCtl(win); err != nil {
		return w, err
	}
	if w.Errors, err = NewErrors(w.ID()); err != nil {
		return w, err
	}

	w.Win = win
	w.Root = toplevel(w.Ctl.Name())
	return w, nil
}

func NewWindowFromEnviron() (*Window, error) {
	winid, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		return nil, err
	}
	return NewWindowFromID(winid)
}

func NewWindowFromID(id int) (*Window, error) {
	win, err := acme.Open(id, nil)
	if err != nil {
		return nil, err
	}

	return NewWindow(win)
}

func NewWindowFromName(name string) (*Window, error) {
	windows, err := acme.Windows()
	if err != nil {
		return nil, err
	}
	name = filepath.Clean(name)

	for _, wi := range windows {
		if filepath.Clean(wi.Name) == name {
			return NewWindowFromID(wi.ID)
		}
	}
	return nil, fmt.Errorf("found no window for name %s", name)
}

func (w *Window) Parent() (*Window, error) {
	return NewWindowFromName(w.Root)
}

func (w *Window) Do(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	cmd.Dir = w.Root
	cmd.Stderr = w.Errors
	cmd.Stdout = w.Errors
	return cmd

}

func (w *Window) FromRoot(name string) string {
	clean, err := filepath.EvalSymlinks(name)
	if err == nil {
		name = clean
	}

	rel, err := filepath.Rel(w.Root, name)
	if err != nil {
		return name
	}
	return rel
}

func (w *Window) Rel() string {
	return w.FromRoot(w.Ctl.Name())
}

// Selection returns the position of the currently selected text in a window.
func (w *Window) Selection() (int, int, error) {
	// First addr read is always 0.
	if _, _, err := w.ReadAddr(); err != nil {
		return 0, 0, err
	}

	if err := w.Win.Ctl("addr=dot\n"); err != nil {
		return 0, 0, err
	}

	q0, q1, err := w.ReadAddr()
	if err != nil {
		return 0, 0, err
	}

	return q0, q1, nil
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

// Ctl is an acme window's ctl file, and provides access to a window's attributes.
type Ctl struct {
	id          int
	tagsize     int
	bodysize    int
	isdirectory bool
	ismodified  bool
	tag         string
	name        string
}

// NewCtl returns a ctl for a window, or an error if the window's ctl cannot be found.
func NewCtl(win *acme.Win) (*Ctl, error) {
	c := &Ctl{}
	ctl, err := win.ReadAll("ctl")
	if err != nil {
		return c, err
	}
	fields := strings.Fields(string(ctl))

	if c.id, err = strconv.Atoi(fields[0]); err != nil {
		return c, err
	}
	if c.tagsize, err = strconv.Atoi(fields[1]); err != nil {
		return c, err
	}
	if c.bodysize, err = strconv.Atoi(fields[2]); err != nil {
		return c, err
	}
	c.isdirectory = fields[3] == "1"
	c.ismodified = fields[4] == "1"

	tag, err := win.ReadAll("tag")
	if err != nil {
		return c, err
	}
	c.tag = string(tag)
	c.name = strings.Fields(c.tag)[0]
	return c, nil
}

func (c *Ctl) IsDirectory() bool {
	return c.isdirectory
}

func (c *Ctl) Name() string {
	return c.name
}

func (c *Ctl) ID() int {
	return c.id
}

func toplevel(path string) string {
	dir := filepath.Dir(path)
	cmd := exec.Command("acme-root")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return dir
	}
	clean := strings.Trim(string(out), "\n")
	delinked, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return clean
	}
	return delinked
}
