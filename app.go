package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Project struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Category string `json:"category"`
}
type ProjectStatus struct {
	Behind   int    `json:"behind"`
	Ahead    int    `json:"ahead"`
	Clean    bool   `json:"clean"`
	HasError bool   `json:"hasError"`
	Error    string `json:"error"`
}
type FileStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}
type Commit struct {
	Hash    string   `json:"hash"`
	Author  string   `json:"author"`
	Date    string   `json:"date"`
	Message string   `json:"message"`
	Refs    string   `json:"refs"`
	Parents []string `json:"parents"`
}
type SyncResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error"`
}
type QuickLink struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	UseCount int    `json:"useCount"`
}
type ContributionDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
type ContributionStats struct {
	Days  []ContributionDay `json:"days"`
	Total int               `json:"total"`
}
type RepoActivity struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Category string `json:"category"`
	Commits  int    `json:"commits"`
}
type Config struct {
	RootDir        string      `json:"rootDir"`
	RecentProjects []string    `json:"recentProjects"`
	QuickLinks     []QuickLink `json:"quickLinks"`
}

type App struct {
	ctx    context.Context
	config Config
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.loadConfig()
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gitmanager", "config.json")
}
func (a *App) loadConfig() {
	data, err := os.ReadFile(configPath())
	if err != nil {
		a.config = Config{QuickLinks: []QuickLink{}, RecentProjects: []string{}}
		return
	}
	_ = json.Unmarshal(data, &a.config)
}
func (a *App) saveConfig() {
	_ = os.MkdirAll(filepath.Dir(configPath()), 0755)
	data, _ := json.MarshalIndent(a.config, "", "  ")
	_ = os.WriteFile(configPath(), data, 0644)
}
func (a *App) GetConfig() Config { return a.config }
func (a *App) SetRootDir(path string) {
	a.config.RootDir = path
	a.saveConfig()
}

func (a *App) SelectDirectory() string {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{Title: "Select Git Projects Root Directory"})
	if err != nil || dir == "" {
		return ""
	}
	a.config.RootDir = dir
	a.saveConfig()
	return dir
}

func (a *App) GetSubfolders(rootDir string) []string {
	if rootDir == "" {
		rootDir = a.config.RootDir
	}
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return []string{}
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

func (a *App) classify(repoPath string) string {
	rootDir := a.config.RootDir
	if rootDir == "" {
		return "Other"
	}
	rel, err := filepath.Rel(rootDir, repoPath)
	if err != nil {
		return "Other"
	}
	parts := strings.SplitN(rel, string(filepath.Separator), 2)
	if len(parts) > 0 && parts[0] != "" && parts[0] != "." {
		return parts[0]
	}
	return "Root"
}

func (a *App) ScanProjects(rootDir string) []Project {
	if rootDir == "" {
		rootDir = a.config.RootDir
	}
	if rootDir == "" {
		return []Project{}
	}
	var mu sync.Mutex
	var projects []Project
	_ = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") && d.Name() != ".git" {
			return filepath.SkipDir
		}
		if d.Name() == ".git" {
			repoPath := filepath.Dir(path)
			mu.Lock()
			projects = append(projects, Project{Name: filepath.Base(repoPath), Path: repoPath, Category: a.classify(repoPath)})
			mu.Unlock()
			return filepath.SkipDir
		}
		return nil
	})
	return projects
}

func (a *App) SaveRecentProject(path string) {
	var recent []string
	for _, p := range a.config.RecentProjects {
		if p != path {
			recent = append(recent, p)
		}
	}
	recent = append([]string{path}, recent...)
	if len(recent) > 10 {
		recent = recent[:10]
	}
	a.config.RecentProjects = recent
	a.saveConfig()
}
func (a *App) GetRecentProjects() []Project {
	var projects []Project
	for _, path := range a.config.RecentProjects {
		if _, err := os.Stat(path); err == nil {
			projects = append(projects, Project{Name: filepath.Base(path), Path: path, Category: a.classify(path)})
		}
	}
	return projects
}

func (a *App) GetQuickLinks() []QuickLink {
	if a.config.QuickLinks == nil {
		return []QuickLink{}
	}
	sorted := make([]QuickLink, len(a.config.QuickLinks))
	copy(sorted, a.config.QuickLinks)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].UseCount > sorted[j].UseCount })
	return sorted
}
func (a *App) AddQuickLink(name, url string) []QuickLink {
	a.config.QuickLinks = append(a.config.QuickLinks, QuickLink{Name: name, URL: url})
	a.saveConfig()
	return a.GetQuickLinks()
}
func (a *App) UseQuickLink(url string) []QuickLink {
	for i := range a.config.QuickLinks {
		if a.config.QuickLinks[i].URL == url {
			a.config.QuickLinks[i].UseCount++
			break
		}
	}
	a.saveConfig()
	return a.GetQuickLinks()
}
func (a *App) RemoveQuickLink(url string) []QuickLink {
	var nl []QuickLink
	for _, l := range a.config.QuickLinks {
		if l.URL != url {
			nl = append(nl, l)
		}
	}
	a.config.QuickLinks = nl
	a.saveConfig()
	return a.GetQuickLinks()
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

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
	return out
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

