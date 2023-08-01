package pty

import (
	"io"

	"golang.org/x/xerrors"
)

// ErrClosed is returned when a PTY is used after it has been closed.
var ErrClosed = xerrors.New("pty: closed")

// PTYCmd is an interface for interacting with a pseudo-TTY where we control
// only one end, and the other end has been passed to a running os.Process.
// nolint:revive
type PTYCmd interface {
	io.Closer

	// Resize sets the size of the PTY.
	Resize(width int, height int) error

	// OutputReader returns an io.Reader for reading the output from the process
	// controlled by the pseudo-TTY
	OutputReader() io.Reader

	// InputWriter returns an io.Writer for writing into to the process
	// controlled by the pseudo-TTY
	InputWriter() io.Writer
}

// PTY is a minimal interface for interacting with pseudo-TTY where this
// process retains access to _both_ ends of the pseudo-TTY (i.e. `ptm` & `pts`
// on Linux).
type PTY interface {
	io.Closer

	// Resize sets the size of the PTY.
	Resize(width int, height int) error

	// Name of the TTY. Example on Linux would be "/dev/pts/1".
	Name() string

	// Output handles TTY output.
	//
	// cmd.SetOutput(pty.Output()) would be used to specify a command
	// uses the output stream for writing.
	//
	// The same stream could be read to validate output.
	Output() io.ReadWriter

	// Input handles TTY input.
	//
	// cmd.SetInput(pty.Input()) would be used to specify a command
	// uses the PTY input for reading.
	//
	// The same stream would be used to provide user input: pty.Input().Write(...)
	Input() io.ReadWriter
}

// Process represents a process running in a PTY.  We need to trigger special processing on the PTY
// on process completion, meaning that we will have goroutines calling Wait() on the process.  Since
// the caller will also typically wait for the process, and it is not safe for multiple goroutines
// to Wait() on a process, this abstraction provides a goroutine-safe interface for interacting with
// the process.
type Process interface {
	// Wait for the command to complete.  Returned error is as for exec.Cmd.Wait()
	Wait() error

	// Kill the command process.  Returned error is as for os.Process.Kill()
	Kill() error
}

// WithFlags represents a PTY whose flags can be inspected, in particular
// to determine whether local echo is enabled.
type WithFlags interface {
	PTY

	// EchoEnabled determines whether local echo is currently enabled for this terminal.
	EchoEnabled() (bool, error)
}

// Controllable represents a PTY that can be controlled via the syscall.RawConn
// interface.
type Controllable interface {
	PTY

	// ControlPTY allows the caller to control the PTY via the syscall.RawConn interface.
	ControlPTY(func(uintptr) error) error

	// ControlTTY allows the caller to control the TTY via the syscall.RawConn interface.
	ControlTTY(func(uintptr) error) error
}

// Options represents a an option for a PTY.
type Option func(*ptyOptions)

type ptyOptions struct {
	setSize bool

	height int
	width  int
}

// WithSize sets the size of the PTY.
func WithSize(width int, height int) Option {
	return func(opts *ptyOptions) {
		opts.setSize = true
		opts.height = height
		opts.width = width
	}
}

// New constructs a new Pty.
func New(opts ...Option) (PTY, error) {
	return newPty(opts...)
}

// readWriter is an implementation of io.ReadWriter that wraps two separate
// underlying file descriptors, one for reading and one for writing, and allows
// them to be accessed separately.
type readWriter struct {
	Reader io.Reader
	Writer io.Writer
}

func (rw readWriter) Read(p []byte) (int, error) {
	return rw.Reader.Read(p)
}

func (rw readWriter) Write(p []byte) (int, error) {
	return rw.Writer.Write(p)
}
