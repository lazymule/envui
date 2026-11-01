# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-11-01

### Added

- Description column: the `#` comment lines directly above a variable (or its trailing `# comment`) are shown in the table; a comment that covers a group of variables is shown, dimmed, for the variables below it too
- Descriptions can be edited in the variable editor and are saved as `#` lines; new variables added from the template or copied from another file take their description along
- Section headings: `# ==== / # Title / # ====` blocks (and `# === Title ===`) are shown as group rows in the table
- Search also matches descriptions and section titles
- Commented-out variables (`# KEY=value`) are shown as "off" rows and can be switched on/off with the "Active" box in the editor; they count as not set in checks and comparisons. Copying a value into a file where the key is only commented out switches it on
- Turkish secret words (`sifre`, `parola`, `anahtar`) and names with a `KEY` / `PWD` / `PASS` part (e.g. `WEB_SERVICE_KEY`) are masked

### Changed

- Edited values are quoted the way dotenv loaders (incl. DelphiMVCFramework) expect: single quotes for values with `#`, `$`, `"` or `\`; backslashes alone (Windows paths) are no longer quoted or escaped
- Deleting a variable also removes its line and its description instead of leaving a blank line
- Names describing a length, duration or limit (`min_parola_uzunlugu`, `TOKEN_EXPIRE_MINUTES`) are no longer treated as secrets

### Fixed

- Editing a value with a backslash (e.g. `logs\analytics`) wrote `"logs\\analytics"`, which some loaders read with a doubled backslash
- Values with `$` were written in double quotes, where loaders may expand `$VARIABLES`

## [0.3.0] - 2026-11-01

### Added

- Compare two env files side by side (Ctrl+D); click a row to copy a value across
- Problem checks: duplicate keys, values with leading/trailing spaces, values that are empty although the template has one
- Warning when env files are not ignored by git, with a one-click "Add to .gitignore"
- Undo / redo (Ctrl+Z, Ctrl+Y / Ctrl+Shift+Z) for every edit, across files
- "Copy to other files" in the variable editor
- Recent projects (header button and start screen) and opening a folder by drag and drop
- Update notice in the status bar (daily check, can be turned off) and "Check for updates" in About
- Portable edition: settings are kept in `envui-portable.json` next to the exe
- Windows installer (`envui-setup.exe`) with Start menu / desktop shortcuts and uninstaller
- Automatic software-OpenGL fallback: the installer and the portable zip include a `softgl` folder (envui.exe + Mesa); envui probes the OpenGL driver at start and switches to it on virtual machines / Remote Desktop. The separate VM package is no longer needed (`ENVUI_SOFTWARE_GL=1` forces it)
- `tools/package_windows.sh` builds the exe, installer and portable zip (replaces `package_vm.sh`)

### Changed

- Developer shown in About is Kemal BAYAT

### Fixed

- Values containing quotes or backslashes gained an extra backslash on every save
- `export KEY=value` lines and trailing `# comments` are now understood and kept
- Lines that are not `KEY=value` are no longer dropped when saving
- Windows (CRLF) line endings are kept
- Unchanged lines are written back byte-for-byte
- Keys that appear twice in a file can now be edited
- Empty secret values show "empty" instead of a mask

## [0.2.0] - 2026-10-31

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
- About dialog (header ⓘ button or F1): version, developer, original author, links, license and the graphics renderer in use
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
