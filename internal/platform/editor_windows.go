package platform

import "os/exec"

// EditorCommand returns an *exec.Cmd that opens path in Notepad.
// Notepad is available on all Windows versions and handles plain-text config
// files without requiring any extra setup.
func EditorCommand(path string) *exec.Cmd {
	return exec.Command("notepad.exe", path)
}
