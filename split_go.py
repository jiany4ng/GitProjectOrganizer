import re

with open("app.go", "r") as f:
    lines = f.readlines()

# Collect start lines of all functions and structs
blocks = []
in_block = False
current_block = []

for i, line in enumerate(lines):
    if line.startswith("package ") or line.startswith("import "):
        # header
        pass
    if line.startswith("type ") and "struct {" in line:
        name = line.split()[1]
        blocks.append({"type": "struct", "name": name, "start": i})
    elif line.startswith("func "):
        m = re.match(r'^func (?:\([^)]+\) )?(\w+)', line)
        if m:
            blocks.append({"type": "func", "name": m.group(1), "start": i})

# Find end lines
for i in range(len(blocks)):
    start = blocks[i]["start"]
    if i < len(blocks) - 1:
        end = blocks[i+1]["start"] - 1
        # walk back empty lines
        while end > start and lines[end].strip() == "":
            end -= 1
        blocks[i]["end"] = end
    else:
        blocks[i]["end"] = len(lines) - 1

files = {
    "models.go": [],
    "config.go": [],
    "scanner.go": [],
    "git.go": [],
    "git_ops.go": [],
    "utils.go": [],
    "app.go": []
}

models_names = ["Project", "ProjectStatus", "FileStatus", "Commit", "SyncResult", "QuickLink", "CommitDetail", "ContributionDay", "ContributionStats", "AppConfig", "Config"]
config_names = ["configPath", "loadConfig", "saveConfig", "GetConfig", "GetQuickLinks", "AddQuickLink", "UseQuickLink", "RemoveQuickLink"]
scanner_names = ["ScanProjects", "GetSubfolders", "classify"]
git_names = ["runGit"]
git_ops_names = ["CheckStatus", "parseCount", "GetBranches", "GetCurrentBranch", "SwitchBranch", "CheckoutCommit", "GetModifiedFiles", "GetCommitHistory", "GetContributionStats", "PullRepo", "CreateBranch", "GetCommitTree", "GetFileAtCommit", "GetFileDiffAtCommit", "SuperSync", "SetBranchColor", "GetBranchColor", "GetPRURL"]
utils_names = ["SelectDirectory", "OpenInVSCode", "OpenTerminal", "GetGitHubURL", "OpenURL"]

for b in blocks:
    name = b["name"]
    content = "".join(lines[b["start"]:b["end"]+1]) + "\n"
    
    if b["type"] == "struct" and name != "App":
        files["models.go"].append(content)
    elif name in config_names:
        files["config.go"].append(content)
    elif name in scanner_names:
        files["scanner.go"].append(content)
    elif name in git_names:
        files["git.go"].append(content)
    elif name in git_ops_names:
        files["git_ops.go"].append(content)
    elif name in utils_names:
        files["utils.go"].append(content)
    else:
        files["app.go"].append(content)

imports = """package main

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

"""

for fname, fcontent in files.items():
    with open(fname, "w") as f:
        f.write(imports)
        f.write("".join(fcontent))

