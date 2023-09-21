package pty

// Resize sets the size of the PTY.
func Resize(f File, width int, height int) error {
	return resize(f, width, height)
}
