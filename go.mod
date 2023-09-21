module github.com/aymanbagabas/go-pty

go 1.20

replace syscall => ../../golang/go/src/syscall

replace golang.org/x/sys => ../../golang/sys

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/charmbracelet/ssh v0.0.0-20230822194956-1a051f898e09
	github.com/creack/pty v1.1.15
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.4
	github.com/u-root/u-root v0.11.0
	golang.org/x/crypto v0.12.0
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691
	golang.org/x/sys v0.12.0
	golang.org/x/term v0.11.0
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
)

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/creack/pty => github.com/aymanbagabas/pty v1.1.19-0.20230803185550-8678d33761d3
