package platform

import "os/exec"

// EditorCommand returns an *exec.Cmd that opens path in the system text editor.
// On Linux this is nano, which works inside a terminal without any special setup.
func EditorCommand(path string) *exec.Cmd {
	return exec.Command("nano", path)
}
