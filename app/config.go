package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

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
