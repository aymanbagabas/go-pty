//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"log"

	"github.com/u-root/u-root/pkg/termios"
	"golang.org/x/crypto/ssh"
)

// terminalModeFlagNames maps the SSH terminal mode flags to mnemonic
// names used by the termios package.
var terminalModeFlagNames = map[uint8]string{
	ssh.VINTR:         "intr",
	ssh.VQUIT:         "quit",
	ssh.VERASE:        "erase",
	ssh.VKILL:         "kill",
	ssh.VEOF:          "eof",
	ssh.VEOL:          "eol",
	ssh.VEOL2:         "eol2",
	ssh.VSTART:        "start",
	ssh.VSTOP:         "stop",
	ssh.VSUSP:         "susp",
	ssh.VDSUSP:        "dsusp",
	ssh.VREPRINT:      "rprnt",
	ssh.VWERASE:       "werase",
	ssh.VLNEXT:        "lnext",
	ssh.VFLUSH:        "flush",
	ssh.VSWTCH:        "swtch",
	ssh.VSTATUS:       "status",
	ssh.VDISCARD:      "discard",
	ssh.IGNPAR:        "ignpar",
	ssh.PARMRK:        "parmrk",
	ssh.INPCK:         "inpck",
	ssh.ISTRIP:        "istrip",
	ssh.INLCR:         "inlcr",
	ssh.IGNCR:         "igncr",
	ssh.ICRNL:         "icrnl",
	ssh.IUCLC:         "iuclc",
	ssh.IXON:          "ixon",
	ssh.IXANY:         "ixany",
	ssh.IXOFF:         "ixoff",
	ssh.IMAXBEL:       "imaxbel",
	ssh.IUTF8:         "iutf8",
	ssh.ISIG:          "isig",
	ssh.ICANON:        "icanon",
	ssh.XCASE:         "xcase",
	ssh.ECHO:          "echo",
	ssh.ECHOE:         "echoe",
	ssh.ECHOK:         "echok",
	ssh.ECHONL:        "echonl",
	ssh.NOFLSH:        "noflsh",
	ssh.TOSTOP:        "tostop",
	ssh.IEXTEN:        "iexten",
	ssh.ECHOCTL:       "echoctl",
	ssh.ECHOKE:        "echoke",
	ssh.PENDIN:        "pendin",
	ssh.OPOST:         "opost",
	ssh.OLCUC:         "olcuc",
	ssh.ONLCR:         "onlcr",
	ssh.OCRNL:         "ocrnl",
	ssh.ONOCR:         "onocr",
	ssh.ONLRET:        "onlret",
	ssh.CS7:           "cs7",
	ssh.CS8:           "cs8",
	ssh.PARENB:        "parenb",
	ssh.PARODD:        "parodd",
	ssh.TTY_OP_ISPEED: "tty_op_ispeed",
	ssh.TTY_OP_OSPEED: "tty_op_ospeed",
}

func applyTerminalModesToFd(fd uintptr, width int, height int, modes ssh.TerminalModes, logger *log.Logger) error {
	if modes == nil {
		modes = ssh.TerminalModes{}
	}

	// Get the current TTY configuration.
	tios, err := termios.GTTY(int(fd))
	if err != nil {
		return fmt.Errorf("GTTY: %w", err)
	}

	// Apply the modes from the SSH request.
	tios.Row = height
	tios.Col = width

	for c, v := range modes {
		if c == ssh.TTY_OP_ISPEED {
			tios.Ispeed = int(v)
			continue
		}
		if c == ssh.TTY_OP_OSPEED {
			tios.Ospeed = int(v)
			continue
		}
		k, ok := terminalModeFlagNames[c]
		if !ok {
			if logger != nil {
				logger.Printf("unknown terminal mode: %d", c)
			}
			continue
		}
		if _, ok := tios.CC[k]; ok {
			tios.CC[k] = uint8(v)
			continue
		}
		if _, ok := tios.Opts[k]; ok {
			tios.Opts[k] = v > 0
			continue
		}

		if logger != nil {
			logger.Printf("unsupported terminal mode: k=%s, c=%d, v=%d", k, c, v)
		}
	}

	// Save the new TTY configuration.
	if _, err := tios.STTY(int(fd)); err != nil {
		return fmt.Errorf("STTY: %w", err)
	}

	return nil
}
