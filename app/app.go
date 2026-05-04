package main

import (
	"context"
	"os"
	"path/filepath"
)

type App struct {
	ctx    context.Context
	config Config
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.loadConfig()
}

func (a *App) SetRootDir(path string) {
	a.config.RootDir = path
	a.saveConfig()
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
