package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (a *App) CheckStatus(path string) ProjectStatus {
	fetchCmd := exec.Command("git", "fetch", "--quiet")
	fetchCmd.Dir = path
	_ = fetchCmd.Run()
	out, err := runGit(path, "status", "-sb")
	if err != nil {
		return ProjectStatus{HasError: true, Error: err.Error()}
	}
	s := ProjectStatus{Clean: true}
	lines := strings.Split(out, "\n")
	if len(lines) > 0 {
		if strings.Contains(lines[0], "ahead") {
			s.Ahead = parseCount(lines[0], "ahead")
		}
		if strings.Contains(lines[0], "behind") {
			s.Behind = parseCount(lines[0], "behind")
		}
	}
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) != "" {
			s.Clean = false
			break
		}
	}
	return s
}

func parseCount(s, keyword string) int {
	idx := strings.Index(s, keyword)
	if idx < 0 {
		return 0
	}
	rest := strings.TrimLeft(s[idx+len(keyword):], " ")
	end := strings.IndexAny(rest, ",]")
	if end > 0 {
		rest = rest[:end]
	}
	n, _ := strconv.Atoi(strings.TrimSpace(rest))
	return n
}

func (a *App) GetBranches(path string) []string {
	out, err := runGit(path, "branch", "--format=%(refname:short)")
	if err != nil || out == "" {
		return []string{}
	}
	var result []string
	for _, b := range strings.Split(out, "\n") {
		if b = strings.TrimSpace(b); b != "" {
			result = append(result, b)
		}
	}
	return result
}

func (a *App) GetCurrentBranch(path string) string {
	out, _ := runGit(path, "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(out)
}

func (a *App) SwitchBranch(path, branch string) string {
	_, err := runGit(path, "checkout", branch)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Sprintf("Error: %s", string(exitErr.Stderr))
		}
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "ok"
}

func (a *App) CheckoutCommit(repoPath, hash string) string {
	_, err := runGit(repoPath, "checkout", hash)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "ok"
}

func (a *App) GetModifiedFiles(path string) []FileStatus {
	out, err := runGit(path, "status", "-s")
	if err != nil || out == "" {
		return []FileStatus{}
	}
	var files []FileStatus
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 3 {
			continue
		}
		files = append(files, FileStatus{Path: strings.TrimSpace(line[3:]), Status: strings.TrimSpace(line[:2])})
	}
	return files
}

func (a *App) GetCommitHistory(path, branch string) []Commit {
	const sep = "|||"
	format := fmt.Sprintf("%%H%s%%an%s%%ar%s%%s%s%%D%s%%P", sep, sep, sep, sep, sep)
	args := []string{"log", "--format=" + format, "--max-count=120"}
	if branch != "" {
		args = append(args, branch)
	}
	out, err := runGit(path, args...)
	if err != nil || out == "" {
		return []Commit{}
	}
	var commits []Commit
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, sep)
		if len(parts) < 6 {
			continue
		}
		var parents []string
		if p := strings.TrimSpace(parts[5]); p != "" {
			parents = strings.Fields(p)
		}
		commits = append(commits, Commit{Hash: parts[0], Author: parts[1], Date: parts[2], Message: parts[3], Refs: parts[4], Parents: parents})
	}
	return commits
}

func (a *App) SuperSync(path string, files []string, message string) SyncResult {
	if len(files) == 0 {
		return SyncResult{Error: "No files selected"}
	}
	if strings.TrimSpace(message) == "" {
		return SyncResult{Error: "Empty commit message"}
	}
	var log strings.Builder
	run := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = path
		out, err := cmd.CombinedOutput()
		log.WriteString(fmt.Sprintf("$ git %s\n%s\n", strings.Join(args, " "), out))
		return err
	}
	if err := run(append([]string{"add", "--"}, files...)...); err != nil {
		return SyncResult{Output: log.String(), Error: "git add failed"}
	}
	if err := run("commit", "-m", message); err != nil {
		return SyncResult{Output: log.String(), Error: "git commit failed"}
	}
	if err := run("push"); err != nil {
		return SyncResult{Output: log.String(), Error: "git push failed"}
	}
	return SyncResult{Success: true, Output: log.String()}
}

// GetContributionStats aggregates daily commit counts for the last 365 days.
// Pass folder="" or "All" to include all repos; otherwise filter by category.

func (a *App) GetContributionStats(folder string) ContributionStats {
	projects := a.ScanProjects(a.config.RootDir)
	if folder != "" && folder != "All" {
		var filtered []Project
		for _, p := range projects {
			if p.Category == folder {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}
	dateCounts := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	since := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	for _, p := range projects {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			
			out, err := runGit(path, "log", "--all", "--format=%cd", "--date=short", "--since="+since)
			if err != nil {
				return
			}
			mu.Lock()
			for _, line := range strings.Split(out, "\n") {
				if l := strings.TrimSpace(line); l != "" {
					dateCounts[l]++
				}
			}
			mu.Unlock()
		}(p.Path)
	}
	wg.Wait()
	total := 0
	start := time.Now().AddDate(-1, 0, 1)
	var days []ContributionDay
	for i := 0; i < 365; i++ {
		d := start.AddDate(0, 0, i)
		ds := d.Format("2006-01-02")
		c := dateCounts[ds]
		total += c
		days = append(days, ContributionDay{Date: ds, Count: c})
	}
	return ContributionStats{Days: days, Total: total}
}
