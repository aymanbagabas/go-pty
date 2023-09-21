//go:build windows
// +build windows

package pty

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

func (c *Cmd) start() (err error) {
	if c.Process != nil {
		return errors.New("already started")
	}

	var pty *conPty
	if c.Stdin == nil || c.Stdout == nil || c.Stderr == nil {
		pty, _, err = open()
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = pty.Close()
			}
		}()
	} else {
		// Try to get the pty from the Stdin, Stdout, or Stderr.
		if p, ok := c.Stdin.(*conPty); ok {
			pty = p
		} else if p, ok := c.Stdout.(*conPty); ok {
			pty = p
		} else if p, ok := c.Stderr.(*conPty); ok {
			pty = p
		} else {
			return errors.New("invalid pty")
		}
	}

	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}

	sys := c.SysProcAttr
	argv0, err := lookExtensions(c.Path, c.Dir)
	if err != nil {
		return err
	}

	if len(c.Dir) != 0 {
		// Windows CreateProcess looks for argv0 relative to the current
		// directory, and, only once the new process is started, it does
		// Chdir(attr.Dir). We are adjusting for that difference here Gby
		// making argv0 absolute.
		var err error
		argv0, err = joinExeDirAndFName(c.Dir, c.Path)
		if err != nil {
			return err
		}
	}

	argv0p, err := windows.UTF16PtrFromString(argv0)
	if err != nil {
		return err
	}

	var cmdline string
	if sys.CmdLine != "" {
		cmdline = sys.CmdLine
	} else {
		cmdline = windows.ComposeCommandLine(c.Args)
	}

	argvp, err := windows.UTF16PtrFromString(cmdline)
	if err != nil {
		return err
	}

	var dirp *uint16
	if len(c.Dir) != 0 {
		dirp, err = windows.UTF16PtrFromString(c.Dir)
		if err != nil {
			return err
		}
	}

	if c.Env == nil {
		c.Env, err = execEnvDefault(sys)
		if err != nil {
			return err
		}
	}

	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()

	siEx := new(windows.StartupInfoEx)
	siEx.Flags = windows.STARTF_USESTDHANDLES
	if sys.HideWindow {
		siEx.Flags |= syscall.STARTF_USESHOWWINDOW
		siEx.ShowWindow = syscall.SW_HIDE
	}

	pi := new(windows.ProcessInformation)

	// Need EXTENDED_STARTUPINFO_PRESENT as we're making use of the attribute list field.
	flags := uint32(windows.CREATE_UNICODE_ENVIRONMENT) | windows.EXTENDED_STARTUPINFO_PRESENT | sys.CreationFlags

	// Allocate an attribute list that's large enough to do the operations we care about
	// 1. Pseudo console setup if one was requested.
	attrList, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return fmt.Errorf("failed to initialize process thread attribute list: %w", err)
	}

	if pty != nil {
		c.sys = conPtySys{
			attrs: attrList,
			pty:   pty,
		}
		if err := pty.UpdateProcThreadAttribute(attrList); err != nil {
			return err
		}
	}

	var zeroSec windows.SecurityAttributes
	pSec := &windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(zeroSec)), InheritHandle: 1}
	tSec := &windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(zeroSec)), InheritHandle: 1}

	siEx.ProcThreadAttributeList = attrList.List() //nolint:govet // unusedwrite: ProcThreadAttributeList will be read in syscall
	siEx.Cb = uint32(unsafe.Sizeof(*siEx))
	if sys.Token != 0 {
		err = windows.CreateProcessAsUser(
			windows.Token(sys.Token),
			argv0p,
			argvp,
			pSec,
			tSec,
			false,
			flags,
			createEnvBlock(addCriticalEnv(dedupEnvCase(true, c.Env))),
			dirp,
			&siEx.StartupInfo,
			pi,
		)
	} else {
		err = windows.CreateProcess(
			argv0p,
			argvp,
			pSec,
			tSec,
			false,
			flags,
			createEnvBlock(addCriticalEnv(dedupEnvCase(true, c.Env))),
			dirp,
			&siEx.StartupInfo,
			pi,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}
	// Don't need the thread handle for anything.
	defer func() {
		_ = windows.CloseHandle(windows.Handle(pi.Thread))
	}()

	// Grab an *os.Process to avoid reinventing the wheel here. The stdlib has great logic around waiting, exit code status/cleanup after a
	// process has been launched.
	c.Process, err = os.FindProcess(int(pi.ProcessId))
	if err != nil {
		// If we can't find the process via os.FindProcess, terminate the process as that's what we rely on for all further operations on the
		// object.
		if tErr := windows.TerminateProcess(pi.Process, 1); tErr != nil {
			return fmt.Errorf("failed to terminate process after process not found: %w", tErr)
		}
		return fmt.Errorf("failed to find process after starting: %w", err)
	}

	if c.ctx != nil {
		go waitOnContext(c)
	}

	return nil
}

type conPtySys struct {
	pty   *conPty
	attrs *windows.ProcThreadAttributeListContainer
}

func (c *Cmd) close() error {
	if c.ProcessState == nil {
		return errors.New("process has not finished")
	}

	sys, ok := c.sys.(conPtySys)
	if !ok {
		return errors.New("invalid type")
	}

	sys.attrs.Delete()
	return sys.pty.Close()
}

func (c *Cmd) wait() (waitErr error) {
	if c.Process == nil {
		return errNotStarted
	}

	if c.waitCalled {
		return errors.New("wait was already called")
	}

	defer func() {
		c.donec <- struct{}{}
		if err := c.close(); err != nil && waitErr == nil {
			waitErr = err
		}
	}()

	c.waitCalled = true
	c.ProcessState, waitErr = c.Process.Wait()
	return waitErr
}

func waitOnContext(c *Cmd) {
	select {
	case <-c.ctx.Done():
		if c.Cancel != nil {
			_ = c.Cancel()
			return
		}
		c.kill()
	case <-c.donec:
		return
	}
}

func execEnvDefault(sys *syscall.SysProcAttr) (env []string, err error) {
	if sys == nil || sys.Token == 0 {
		return syscall.Environ(), nil
	}

	var block *uint16
	err = windows.CreateEnvironmentBlock(&block, windows.Token(sys.Token), false)
	if err != nil {
		return nil, err
	}

	defer windows.DestroyEnvironmentBlock(block)
	blockp := uintptr(unsafe.Pointer(block))

	for {

		// find NUL terminator
		end := unsafe.Pointer(blockp)
		for *(*uint16)(end) != 0 {
			end = unsafe.Pointer(uintptr(end) + 2)
		}

		n := (uintptr(end) - uintptr(unsafe.Pointer(blockp))) / 2
		if n == 0 {
			// environment block ends with empty string
			break
		}

		entry := (*[(1 << 30) - 1]uint16)(unsafe.Pointer(blockp))[:n:n]
		env = append(env, string(utf16.Decode(entry)))
		blockp += 2 * (uintptr(len(entry)) + 1)
	}
	return
}

//
// Below are a bunch of helpers for working with Windows' CreateProcess family of functions. These are mostly exact copies of the same utilities
// found in the go stdlib.
//

func lookExtensions(path, dir string) (string, error) {
	if filepath.Base(path) == path {
		path = filepath.Join(".", path)
	}

	if dir == "" {
		return exec.LookPath(path)
	}

	if filepath.VolumeName(path) != "" {
		return exec.LookPath(path)
	}

	if len(path) > 1 && os.IsPathSeparator(path[0]) {
		return exec.LookPath(path)
	}

	dirandpath := filepath.Join(dir, path)

	// We assume that LookPath will only add file extension.
	lp, err := exec.LookPath(dirandpath)
	if err != nil {
		return "", err
	}

	ext := strings.TrimPrefix(lp, dirandpath)

	return path + ext, nil
}

func isSlash(c uint8) bool {
	return c == '\\' || c == '/'
}

func normalizeDir(dir string) (name string, err error) {
	ndir, err := syscall.FullPath(dir)
	if err != nil {
		return "", err
	}
	if len(ndir) > 2 && isSlash(ndir[0]) && isSlash(ndir[1]) {
		// dir cannot have \\server\share\path form
		return "", syscall.EINVAL
	}
	return ndir, nil
}

func volToUpper(ch int) int {
	if 'a' <= ch && ch <= 'z' {
		ch += 'A' - 'a'
	}
	return ch
}

func joinExeDirAndFName(dir, p string) (name string, err error) {
	if len(p) == 0 {
		return "", syscall.EINVAL
	}
	if len(p) > 2 && isSlash(p[0]) && isSlash(p[1]) {
		// \\server\share\path form
		return p, nil
	}
	if len(p) > 1 && p[1] == ':' {
		// has drive letter
		if len(p) == 2 {
			return "", syscall.EINVAL
		}
		if isSlash(p[2]) {
			return p, nil
		} else {
			d, err := normalizeDir(dir)
			if err != nil {
				return "", err
			}
			if volToUpper(int(p[0])) == volToUpper(int(d[0])) {
				return syscall.FullPath(d + "\\" + p[2:])
			} else {
				return syscall.FullPath(p)
			}
		}
	} else {
		// no drive letter
		d, err := normalizeDir(dir)
		if err != nil {
			return "", err
		}
		if isSlash(p[0]) {
			return windows.FullPath(d[:2] + p)
		} else {
			return windows.FullPath(d + "\\" + p)
		}
	}
}

// createEnvBlock converts an array of environment strings into
// the representation required by CreateProcess: a sequence of NUL
// terminated strings followed by a nil.
// Last bytes are two UCS-2 NULs, or four NUL bytes.
func createEnvBlock(envv []string) *uint16 {
	if len(envv) == 0 {
		return &utf16.Encode([]rune("\x00\x00"))[0]
	}
	length := 0
	for _, s := range envv {
		length += len(s) + 1
	}
	length += 1

	b := make([]byte, length)
	i := 0
	for _, s := range envv {
		l := len(s)
		copy(b[i:i+l], []byte(s))
		copy(b[i+l:i+l+1], []byte{0})
		i = i + l + 1
	}
	copy(b[i:i+1], []byte{0})

	return &utf16.Encode([]rune(string(b)))[0]
}

// dedupEnvCase is dedupEnv with a case option for testing.
// If caseInsensitive is true, the case of keys is ignored.
func dedupEnvCase(caseInsensitive bool, env []string) []string {
	out := make([]string, 0, len(env))
	saw := make(map[string]int, len(env)) // key => index into out
	for _, kv := range env {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			out = append(out, kv)
			continue
		}
		k := kv[:eq]
		if caseInsensitive {
			k = strings.ToLower(k)
		}
		if dupIdx, isDup := saw[k]; isDup {
			out[dupIdx] = kv
			continue
		}
		saw[k] = len(out)
		out = append(out, kv)
	}
	return out
}

// addCriticalEnv adds any critical environment variables that are required
// (or at least almost always required) on the operating system.
// Currently this is only used for Windows.
func addCriticalEnv(env []string) []string {
	for _, kv := range env {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			continue
		}
		k := kv[:eq]
		if strings.EqualFold(k, "SYSTEMROOT") {
			// We already have it.
			return env
		}
	}
	return append(env, "SYSTEMROOT="+os.Getenv("SYSTEMROOT"))
}
