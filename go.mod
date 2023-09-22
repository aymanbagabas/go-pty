module github.com/aymanbagabas/go-pty

go 1.20

// The replace fork includes these changes:
// - https://github.com/creack/pty/pull/168: Use upstream compiler for linux/riscv64 and freebsd/riscv64
// - https://github.com/creack/pty/pull/167: Avoid calls to (*os.File).Fd() and operations on raw file descriptor ints
replace github.com/creack/pty => github.com/aymanbagabas/pty v1.1.19-0.20230922024246-7bc6991e768a

require (
	github.com/creack/pty v1.1.15
	github.com/u-root/u-root v0.11.0
	golang.org/x/crypto v0.12.0
	golang.org/x/sys v0.12.0
)
