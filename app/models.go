package main

type Project struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Category string `json:"category"`
}

type ProjectStatus struct {
	Branch      string `json:"branch"`
	BranchColor string `json:"branchColor"`
	Behind      int    `json:"behind"`
	Ahead       int    `json:"ahead"`
	Clean       bool   `json:"clean"`
	HasError    bool   `json:"hasError"`
	Error       string `json:"error"`
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
	RootDir        string            `json:"rootDir"`
	RecentProjects []string          `json:"recentProjects"`
	QuickLinks     []QuickLink       `json:"quickLinks"`
	BranchColors   map[string]string `json:"branchColors"`
}
