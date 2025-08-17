package browser

import (
	"os"
	"os/exec"
	"runtime"
)

// findChromeExecutable locates Chrome/Chromium/Brave executable
func findChromeExecutable() string {
	var candidates []string

	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}
	case "linux":
		candidates = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
			"/usr/bin/brave-browser",
		}
	case "windows":
		candidates = []string{
			"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files\\Chromium\\Application\\chromium.exe",
			"C:\\Program Files (x86)\\Chromium\\Application\\chromium.exe",
		}
	}

	// Check each candidate
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Try to find in PATH
	if path, err := exec.LookPath("google-chrome"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chromium"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chromium-browser"); err == nil {
		return path
	}
	if path, err := exec.LookPath("brave-browser"); err == nil {
		return path
	}

	return ""
}
