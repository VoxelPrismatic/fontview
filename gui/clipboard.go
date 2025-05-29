package gui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func copyFunc() (string, error) {
	switch runtime.GOOS {
	case "linux":
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "WAYLAND") {
				_, err := exec.LookPath("wl-copy")
				if err != nil {
					return "", fmt.Errorf("wl-copy: %s", err.Error())
				} else {
					return "echo -n \"%s\" | wl-copy", nil
				}
			}
		}
		_, err := exec.LookPath("xclip")
		if err != nil {
			return "", fmt.Errorf("xsel: %s", err.Error())
		}
		return "echo -n \"%s\" | xclip", nil
	}
	return "", fmt.Errorf("%s not supported", runtime.GOOS)
}

var copyCmd, copyErr = copyFunc()

func copyRune() error {
	if copyErr != nil {
		return copyErr
	}

	lines := strings.Split(info_CodeLabel.Text(), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("nothing to copy")
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf(copyCmd, lines[0]))
	return cmd.Run()
}
