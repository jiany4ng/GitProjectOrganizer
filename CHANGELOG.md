# Changelog

All notable changes to this project will be documented in this file.

## [1.1.1] - 2026-05-04

### Added
- **UI**: Display active branch name directly on project cards in the dashboard.
- **UI**: Visual branch color accents on dashboard cards (border + gradient + glow).
- **Backend**: Project-specific branch color persistence (stored as `repoPath:branch`).

### Fixed
- **Git**: Fixed critical pathspec bug ("pp.go") by preserving leading whitespace in `git status` output.
- **Go**: Resolved unused import error in `git.go`.
- **UI**: Improved real-time UI synchronization when updating branch colors from the detail panel.

### Changed
- **Architecture**: Modularized codebase moved to `app/` directory for better organization.
- **UI**: Strengthened branch color visuals with thicker borders and subtle glow effects.

## [1.1.0] - 2026-05-04

### Added
- **Backend**: New `git_extended.go` module implementing core Git operations: `PullRepo`, `CreateBranch`, `GetPRURL`.
- **Backend**: Added `BranchColors` persistence in `models.go` and corresponding getters/setters.
- **Backend**: Advanced commit exploration methods: `GetCommitTree`, `GetFileAtCommit`, and `GetFileDiffAtCommit`.
- **UI**: Hierarchical Tree View in Commit Explorer for better directory navigation.
- **UI**: Smart file filtering in Commit Explorer (hiding binaries, images, and non-code files).
- **UI**: Quick "Pull" buttons added directly to project cards in the dashboard.
- **UI**: Added `--wails-drop-target` support for window dragging from the top bar and sidebar header.
- **UI**: Interactive tooltips for the Contribution Grid showing daily activity details.

### Fixed
- **Frontend**: Resolved critical `SyntaxError` and initialization blocks caused by missing backend bindings.
- **Scanner**: Improved `ScanProjects` performance and accuracy by skipping `node_modules`, `vendor`, `build`, and `dist` directories.
- **UX**: Centered dashboard layout with `max-width` constraints for improved readability on ultra-wide screens.
- **UX**: Fixed tooltip visibility by resolving CSS class conflicts (`hidden` vs `visible`).
- **Git**: Improved error reporting in `Pull` operations to display actual Git stderr output.

### Changed
- **UI**: Restructured Repository Detail panel to prioritize tools (Terminal, GitHub, VSCode) and group branch actions logically.
