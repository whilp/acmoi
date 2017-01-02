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

// Window wraps acme.Win.
type Window struct {
	*acme.Win
	*Ctl
	Errors *Errors

	// The root of the project that contains the file or directory being
	// edited by the window.
	Root string
}

// NewWindow creates a new Window, wrapping an existing acme.Win. Each Window is associated with a Ctl and Errors. The Root for the Window is determined by calling acme-root in the directory that contains the file or directory shown in the window.
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
	w.Root = toplevel(w.Ctl.Name()) + "/"
	return w, nil
}

// NewWindowFromEnviron creates a new Window based on the winid environment variable. If acme does not know about a window with this ID, error will be set.
func NewWindowFromEnviron() (*Window, error) {
	winid, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		return nil, err
	}
	return NewWindowFromID(winid)
}

// NewWindowFromID creates a new Window given a window ID. If acme does not know about a window with this ID, error will be set.
func NewWindowFromID(id int) (*Window, error) {
	win, err := acme.Open(id, nil)
	if err != nil {
		return nil, err
	}

	return NewWindow(win)
}

// NewWindowFromName creates a new Window given the name of the file or directory shown in that window. If acme does not know about a window with this name, error will be set unless create is true. If create is true, a new window will be created in acme and returned.
func NewWindowFromName(name string, create bool) (*Window, error) {
	windows, err := acme.Windows()
	if err != nil {
		return nil, err
	}

	for _, wi := range windows {
		if wi.Name == name {
			return NewWindowFromID(wi.ID)
		}
	}

	if create {
		w, err := acme.New()
		if err != nil {
			return nil, err
		}
		w.Name(name)
		return NewWindow(w)
	}
	return nil, fmt.Errorf("window for name does not exist: %s", name)
}

// Parent finds the window showing the .guide file at the root of the project. The root is typically determined by acme-root (see NewWindow).
func (w *Window) Parent() (*Window, error) {
	name := filepath.Join(w.Root, ".guide")
	return NewWindowFromName(name, true)
}

// Do runs a command in the root directory of the project. The commands output is routed to the window's Errors file for display in acme.
func (w *Window) Do(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	cmd.Dir = w.Root
	cmd.Stderr = w.Errors
	cmd.Stdout = w.Errors
	return cmd
}

// FromRoot returns a path to a file relative to the project's root directory.
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

// Rel returns the window's name relative to the project's root directory.
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
	*client.Fid
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
	return &Cons{Fid: f}, nil
}

// Errors is a special file associated with each acme window for logging errors.
type Errors struct {
	*client.Fid
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

	return &Errors{Fid: f}, nil
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

// IsDirectory returns true if the isdirectory field for this ctl's window is set to 1.
func (c *Ctl) IsDirectory() bool {
	return c.isdirectory
}

// Name returns the name of this ctl's window.
func (c *Ctl) Name() string {
	return c.name
}

// ID returns the ID of this ctl's window.
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
