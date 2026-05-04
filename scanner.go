package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

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
		if d.Name() == "node_modules" || d.Name() == "vendor" || d.Name() == "build" || d.Name() == "dist" {
			return filepath.SkipDir
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
