# envui

A visual .env file editor built with Go and Fyne.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

## Why?

Managing .env files sucks:

- Which vars are missing vs `.env.example`?
- What's inherited from shell vs explicit?
- Did I accidentally commit a secret?
- What's different between local and staging?

envui gives you a visual editor to manage environment variables across multiple .env files.

![envui dark theme](docs/screenshot-dark.png)

## Features

- **Visual Editor** - Table view of all variables, click a row to edit
- **Secret Masking** - Auto-detect and mask sensitive values (API_KEY, PASSWORD, TOKEN, etc.)
- **Multi-file Support** - Switch between `.env`, `.env.local`, `.env.staging`, etc.
- **Missing Variables** - Keys in `.env.example` that aren't set are highlighted; click one to add it
- **Source Tracking** - See which variables come from file vs shell environment
- **Search** - Filter variables by name or value
- **Light & Dark Themes** - Toggle from the header; your choice is remembered
- **Languages** - English and Turkish (Türkçe); picked from your system language and changeable in Settings
- **Comment Preservation** - Your comments stay intact when saving

### Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| `Ctrl+S` | Save all changed files |
| `Ctrl+N` | Add a variable |
| `Ctrl+F` | Search |
| `Ctrl+O` | Open a folder |

(`Cmd` instead of `Ctrl` on macOS.)

## Installation

### From Source

```bash
git clone https://github.com/mrf/envui.git
cd envui
go build -o envui .
```

### Requirements

- Go 1.21 or later
- For building on Linux, you may need: `libgl1-mesa-dev xorg-dev`

### Building for Windows

Natively (with a C compiler such as [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MSYS2 on the PATH):

```bash
go build -ldflags "-H=windowsgui" -o envui.exe .
```

Cross-compiling from Linux (needs `gcc-mingw-w64-x86-64`):

```bash
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "-H=windowsgui -s -w" -o envui.exe .
```

## Usage

```bash
# Open the app and select a folder
./envui

# Or open a specific project directory
./envui ~/Projects/my-app
```

### Keyboard Shortcuts

- Click on any variable row to edit
- Use the "Show Secrets" checkbox to reveal masked values

## Supported Files

envui automatically discovers these files:

- `.env`
- `.env.local`
- `.env.development`
- `.env.staging`
- `.env.production`
- `.env.example`
- `.env.sample`
- `.env.template`
- `.env.test`

## Secret Detection

Variables matching these patterns are automatically masked:

- `PASSWORD`, `SECRET`, `TOKEN`
- `API_KEY`, `APIKEY`, `API_SECRET`
- `PRIVATE_KEY`, `ACCESS_KEY`
- `AUTH`, `CREDENTIAL`, `DSN`

## Building

```bash
# Development
go run .

# Build binary
go build -o envui .

# Run tests
go test ./...

# macOS app bundle (optional)
fyne package -os darwin -icon icon.png
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
