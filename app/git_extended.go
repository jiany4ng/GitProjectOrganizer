package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func (a *App) PullRepo(path string) SyncResult {
	cmd := exec.Command("git", "pull")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		errorMessage := string(out)
		if errorMessage == "" {
			errorMessage = err.Error()
		}
		return SyncResult{Error: errorMessage, Output: string(out)}
	}
	return SyncResult{Success: true, Output: string(out)}
}

func (a *App) CreateBranch(path, name string) string {
	_, err := runGit(path, "checkout", "-b", name)
	if err != nil {
		return err.Error()
	}
	return "ok"
}

func (a *App) GetPRURL(path, head, base string) string {
	repoURL := a.GetGitHubURL(path)
	if repoURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/compare/%s...%s", repoURL, base, head)
}

func (a *App) SetBranchColor(repoPath, branch, color string) error {
	if a.config.BranchColors == nil {
		a.config.BranchColors = make(map[string]string)
	}
	a.config.BranchColors[repoPath+":"+branch] = color
	a.saveConfig()
	return nil
}

func (a *App) GetBranchColor(repoPath, branch string) string {
	if a.config.BranchColors == nil {
		return ""
	}
	return a.config.BranchColors[repoPath+":"+branch]
}

func (a *App) GetCommitTree(repoPath, hash string) []string {
	out, err := runGit(repoPath, "show", "--name-only", "--pretty=format:", hash)
	out = strings.TrimSpace(out)
	if err != nil || out == "" {
		return []string{}
	}
	var files []string
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			files = append(files, l)
		}
	}
	return files
}

func (a *App) GetFileAtCommit(repoPath, hash, file string) string {
	out, err := runGit(repoPath, "show", fmt.Sprintf("%s:%s", hash, file))
	if err != nil {
		return err.Error()
	}
	return out
}

func (a *App) GetFileDiffAtCommit(repoPath, hash, file string) string {
	// Try standard diff against parent
	out, err := runGit(repoPath, "diff", hash+"^", hash, "--", file)
	if err != nil || out == "" {
		// Fallback to git show if there is no parent or other issue
		out2, err2 := runGit(repoPath, "show", hash, "--", file)
		if err2 != nil {
			return err2.Error()
		}
		return out2
	}
	return out
}
