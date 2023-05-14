package compression

import (
	"os"
	"path/filepath"
	"runtime"
)

func SetupWindowsTests() {
	if runtime.GOOS == "windows" {
		path := os.Getenv("PATH")
		tools_win_dir, _ := filepath.Abs("../../tools_win")
		os.Setenv("PATH", path+";"+tools_win_dir)
	}
}
