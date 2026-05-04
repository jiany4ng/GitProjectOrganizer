package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) SelectDirectory() string {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{Title: "Select Git Projects Root Directory"})
	if err != nil || dir == "" {
		return ""
	}
	a.config.RootDir = dir
	a.saveConfig()
	return dir
}

func (a *App) OpenInVSCode(path string) string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("code"); err == nil {
			cmd = exec.Command("code", path)
		} else {
			cmd = exec.Command("open", "-a", "Visual Studio Code", path)
		}
	case "windows":
		cmd = exec.Command("cmd", "/C", "code", path)
	default:
		cmd = exec.Command("code", path)
	}
	if err := cmd.Start(); err != nil {
		return err.Error()
	}
	return "ok"
}

// OpenTerminal opens a terminal emulator at the given directory (cross-platform).

func (a *App) OpenTerminal(path string) string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		// Use AppleScript to open Terminal.app at the given path
		script := fmt.Sprintf(`tell application "Terminal"
	do script "cd %s"
	activate
end tell`, strings.ReplaceAll(path, `"`, `\"`))
		cmd = exec.Command("osascript", "-e", script)
	case "windows":
		// Try Windows Terminal first, fall back to cmd
		if _, err := exec.LookPath("wt"); err == nil {
			cmd = exec.Command("wt", "-d", path)
		} else {
			cmd = exec.Command("cmd", "/C", "start", "cmd", "/K", "cd", "/d", path)
		}
	default: // Linux
		type termInfo struct {
			bin  string
			args []string
		}
		terminals := []termInfo{
			{"gnome-terminal", []string{"--working-directory=" + path}},
			{"konsole", []string{"--workdir", path}},
			{"xfce4-terminal", []string{"--working-directory=" + path}},
			{"tilix", []string{"-d", path}},
			{"xterm", []string{"-e", fmt.Sprintf("bash -c 'cd %q; exec $SHELL'", path)}},
		}
		for _, t := range terminals {
			if _, err := exec.LookPath(t.bin); err == nil {
				cmd = exec.Command(t.bin, t.args...)
				break
			}
		}
		if cmd == nil {
			return "No supported terminal emulator found"
		}
	}
	if err := cmd.Start(); err != nil {
		return err.Error()
	}
	return "ok"
}

func (a *App) GetGitHubURL(path string) string {
	out, err := runGit(path, "remote", "get-url", "origin")
	if err != nil || out == "" {
		return ""
	}
	url := strings.TrimSpace(out)
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimSuffix(strings.TrimPrefix(url, "git@github.com:"), ".git")
		return "https://github.com/" + url
	}
	if strings.Contains(url, "github.com") {
		return strings.TrimSuffix(url, ".git")
	}
	return ""
}

func (a *App) OpenURL(url string) string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return err.Error()
	}
	return "ok"
}
