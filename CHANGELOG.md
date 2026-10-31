# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Refreshed UI: custom light/dark theme (remembered between runs), header bar, file sidebar and status bar
- Search box to filter variables by name or value
- Variables listed in `.env.example` / `.env.sample` / `.env.template` but missing from the current file are shown and can be added with one click
- Optional display of shell environment variables
- "New .env file" action (optionally copying a template) and an empty state when a folder has no env files
- Reload from disk, copy value to clipboard, keyboard shortcuts (Ctrl+S / N / F / O)
- Unsaved-changes prompt when quitting, reloading or opening another folder
- Turkish translation and language support; language follows the system locale by default
- Settings dialog: language, theme and startup options (show shell variables / secrets)
- Virtual machine / Remote Desktop support: when Mesa's software OpenGL (opengl32.dll + libgallium_wgl.dll) is next to envui.exe, envui uses the llvmpipe CPU renderer; `tools/package_vm.sh` builds a ready-to-use zip
- Application icon (window, taskbar and embedded in envui.exe with version info)
- Colour palette based on the GLARK brand blues (light and dark themes)
- Folder picker remembers the last opened project and starts next to it
- Turkish README (default) with an English version in README.en.md; credits to the original author

### Fixed

- Table header columns now line up with the table
- Clicking the same row twice now reopens the editor
- Deleting a variable closes the edit dialog and asks for confirmation
- Long or non-ASCII values (e.g. Turkish characters) are no longer cut off incorrectly
- Invalid or duplicate variable names are rejected

## [0.1.0] - 2025-12-11

### Added

- Initial release
- Visual table editor for .env files
- Automatic secret masking for sensitive variables (PASSWORD, TOKEN, API_KEY, etc.)
- Multi-file support (.env, .env.local, .env.staging, etc.)
- Source tracking (file vs shell environment)
- Comment preservation when saving
- Add, edit, and delete variables
- Cross-platform support (macOS, Linux)
