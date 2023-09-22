module examples

go 1.20

replace github.com/aymanbagabas/go-pty => ../

replace github.com/creack/pty => github.com/aymanbagabas/pty v1.1.19-0.20230922024246-7bc6991e768a

require (
	github.com/aymanbagabas/go-pty v0.0.0-00010101000000-000000000000
	github.com/charmbracelet/ssh v0.0.0-20230822194956-1a051f898e09
	github.com/u-root/u-root v0.11.0
	golang.org/x/crypto v0.13.0
	golang.org/x/term v0.12.0
)

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/creack/pty v1.1.15 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
