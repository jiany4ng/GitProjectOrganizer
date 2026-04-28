# ⎇ Git Project Manager

A lightweight, portable desktop application to visualize and manage your local Git repositories — no API tokens required.

Built with **Go + Wails v2** (backend) and **Vanilla HTML/CSS/JS** (frontend).

---

## Features

- **Dashboard** with world clocks (Spain, UK, US, Australia)
- **GitHub-style contribution heatmap** filterable by folder
- **Repository browser** — auto-discovers all `.git` repos under a root directory
- **Sidebar** navigation built from your actual subfolder structure
- **Commit history** visualization (gitgraph.js)
- **Branch switcher** and **commit checkout**
- **Super Sync** — stage files, write message, push in one click
- **Open in VS Code** and **Open Terminal** at any repo path
- **Open in GitHub** — auto-detects remote URL
- **Quick Links** cards with usage-frequency sorting
- **Recent repositories** tracking

---

## Prerequisites

You need the following tools installed on your system regardless of OS:

| Tool | Version | Download |
|------|---------|----------|
| **Go** | 1.18+ | https://go.dev/dl/ |
| **Node.js** | 16+ | https://nodejs.org/ |
| **Wails CLI** | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| **Git** | any | https://git-scm.com/ |

Verify everything is installed:

```bash
go version
node --version
wails doctor
git --version
```

---

## macOS

### Install dependencies

```bash
# 1. Install Go (via Homebrew or from go.dev)
brew install go

# 2. Install Node.js
brew install node

# 3. Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 4. Add Go bin to PATH (add to ~/.zshrc or ~/.bash_profile)
export PATH=$PATH:$(go env GOPATH)/bin
```

> **Xcode Command Line Tools** are required. Install with:
> ```bash
> xcode-select --install
> ```

### Run in development mode (hot reload)

```bash
cd gitmanager
wails dev
```

The app window opens automatically. You can also open **http://localhost:34115** in your browser.

### Build a standalone `.app`

```bash
wails build
```

Output: `build/bin/gitmanager.app`

Double-click to run, or distribute the `.app` bundle.

> **Note:** macOS may show a security warning on first launch for unsigned apps.  
> Go to **System Settings → Privacy & Security → Open Anyway**, or run:
> ```bash
> xattr -cr build/bin/gitmanager.app
> ```

### Terminal integration

The "Open Terminal" button uses **AppleScript** to open **Terminal.app** at the project path. If you use **iTerm2**, it will also work since iTerm2 registers itself as the default terminal handler in some setups.

---

## Windows

### Install dependencies

1. **Go**: Download from https://go.dev/dl/ — run the `.msi` installer
2. **Node.js**: Download from https://nodejs.org/ — run the `.msi` installer
3. **Git**: Download from https://git-scm.com/download/win
4. **WebView2 Runtime**: Usually pre-installed on Windows 10/11. If not: https://developer.microsoft.com/en-us/microsoft-edge/webview2/
5. **Wails CLI**: Open **Command Prompt** or **PowerShell** and run:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify:
```powershell
wails doctor
```

### Run in development mode

```powershell
cd gitmanager
wails dev
```

### Build a standalone `.exe`

```powershell
wails build
```

Output: `build\bin\gitmanager.exe`

The executable is self-contained — no installer needed. You can place it anywhere.

### Terminal integration

The "Open Terminal" button will:
1. Try **Windows Terminal** (`wt -d <path>`) if installed
2. Fall back to **cmd.exe** opening at the project directory

Install Windows Terminal from the Microsoft Store for the best experience.

---

## Linux

### Install dependencies

**Ubuntu / Debian:**
```bash
# Go
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# Node.js (via NodeSource)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# WebKit / GTK (required by Wails on Linux)
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev

# Git
sudo apt-get install -y git

# Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**Fedora / RHEL:**
```bash
sudo dnf install golang nodejs webkit2gtk3-devel gtk3-devel git
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**Arch Linux:**
```bash
sudo pacman -S go nodejs webkit2gtk gtk3 git
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify:
```bash
wails doctor
```

### Run in development mode

```bash
cd gitmanager
wails dev
```

### Build a standalone binary

```bash
wails build
```

Output: `build/bin/gitmanager`

Make it executable and run:
```bash
chmod +x build/bin/gitmanager
./build/bin/gitmanager
```

### Terminal integration

The "Open Terminal" button tries the following terminal emulators **in order**:

1. `gnome-terminal`
2. `konsole`
3. `xfce4-terminal`
4. `tilix`
5. `xterm`

Install at least one:
```bash
# GNOME
sudo apt install gnome-terminal

# KDE
sudo apt install konsole

# XFCE
sudo apt install xfce4-terminal
```

---

## VS Code Integration

The "Open in VS Code" button requires the `code` CLI to be available in your `$PATH`.

- **macOS**: In VS Code, open the Command Palette (`⇧⌘P`) → **Shell Command: Install 'code' command in PATH**
- **Windows**: The `code` CLI is added to PATH automatically during installation
- **Linux**: Install via your package manager or from https://code.visualstudio.com/

---

## Project Structure

```
gitmanager/
├── app.go              # Go backend — scanning, git commands, file ops
├── main.go             # Wails app entry point
├── wails.json          # Wails configuration
├── go.mod / go.sum     # Go module files
└── frontend/
    ├── index.html      # App shell
    ├── style.css       # Light theme styles
    ├── main.js         # Frontend logic
    └── wailsjs/        # Auto-generated Wails JS bindings
```

---

## Config File

User preferences (root directory, quick links, recent repos) are saved to:

| OS | Path |
|----|------|
| macOS / Linux | `~/.gitmanager/config.json` |
| Windows | `%USERPROFILE%\.gitmanager\config.json` |

Delete this file to reset all settings.

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `wails: command not found` | Run `export PATH=$PATH:$(go env GOPATH)/bin` and add it to your shell profile |
| App won't open on macOS | Run `xattr -cr build/bin/gitmanager.app` to remove quarantine flag |
| White/blank window | WebView2 not installed (Windows) — download from Microsoft |
| `webkit2gtk` errors on Linux | Install `libwebkit2gtk-4.0-dev` (Debian) or `webkit2gtk3-devel` (Fedora) |
| Terminal button does nothing | Install a supported terminal emulator (see Linux section above) |
| Contribution grid is slow | Normal for large workspaces — it runs `git log` on every repo concurrently |

---

## License

MIT
