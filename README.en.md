# envui

[Türkçe](README.md) | **English**

A visual .env file editor built with Go and Fyne.

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

## Why?

Managing .env files is a pain:

- Which variables are missing compared to `.env.example`?
- Which ones are set in the file, and which come from the shell?
- Did I accidentally commit a secret?
- What's different between local and staging?

envui lets you view and edit environment variables across multiple .env files from a single window.

![envui dark theme](docs/screenshot-dark.png)

## Features

- **Visual editor**: All variables are listed in a table. Click a row to edit it.
- **Secret masking**: Sensitive values such as API_KEY, PASSWORD and TOKEN are detected and masked automatically.
- **Multi-file support**: Switch between `.env`, `.env.local`, `.env.staging` and similar files.
- **Missing variables**: Keys in `.env.example` that aren't set in your file are highlighted. Click one to add it with the example value.
- **Source tracking**: See whether a variable comes from the file or the shell environment.
- **Search**: Filter variables by name or value.
- **Light and dark themes**: Your choice is remembered between runs.
- **Languages**: Turkish and English. Picked from your system language and changeable in Settings.
- **Remembers folders**: The folder picker opens next to the project you opened last.
- **Comment preservation**: Comments in your files stay intact when saving.

### Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| `Ctrl+S` | Save all changed files |
| `Ctrl+N` | Add a variable |
| `Ctrl+F` | Search |
| `Ctrl+O` | Open a folder |

Use `Cmd` instead of `Ctrl` on macOS.

## Installation

### From source

```bash
git clone https://github.com/lazymule/envui.git
cd envui
go build -o envui .
```

### Requirements

- Go 1.23 or later
- A C compiler (Fyne uses cgo)
- On Linux you may also need: `libgl1-mesa-dev xorg-dev`

### Building for Windows

To build natively on Windows, you need a C compiler on the PATH (such as [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MSYS2):

```bash
go build -ldflags "-H=windowsgui" -o envui.exe .
```

To cross-compile from Linux, you need `gcc-mingw-w64-x86-64`:

```bash
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "-H=windowsgui -s -w" -o envui.exe .
```

## Usage

Open the app and pick a folder:

```bash
./envui
```

Or open a project folder directly:

```bash
./envui ~/Projects/my-app
```

On Windows, just double-click `envui.exe`.

## Settings

Open Settings with the ⚙ icon in the header:

- **Language**: Türkçe or English. The interface updates immediately, and the open project and unsaved changes are kept.
- **Theme**: Light or Dark.
- **On start**: Whether shell variables and secret values are shown on start.

The folder picker is a built-in Fyne component, so it is always shown in English.

## Supported files

envui automatically discovers these files:

- `.env`
- `.env.local`
- `.env.development`
- `.env.staging`
- `.env.production`
- `.env.prod`
- `.env.example`
- `.env.sample`
- `.env.template`
- `.env.test`

`.env.example`, `.env.sample` and `.env.template` are treated as templates. Missing variables are worked out from them.

## Secret detection

Variables whose names contain these patterns are masked automatically:

- `PASSWORD`, `SECRET`, `TOKEN`
- `API_KEY`, `APIKEY`, `API_SECRET`
- `PRIVATE_KEY`, `ACCESS_KEY`
- `AUTH`, `CREDENTIAL`, `DSN`

## Development

```bash
# Run
go run .

# Build
go build -o envui .

# Run tests
go test ./...
```

To add a language, add a section to the `messages` table in `ui/i18n.go` and an entry to the `languages` list. The tests flag any missing translations.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Acknowledgements

This project is a fork of [mrf/envui](https://github.com/mrf/envui) by **Mark Ferree**. The `.env` parser, the comment-preserving writer and the original interface are his work. Many thanks to him for sharing the project as open source. 🙏

This fork adds a refreshed interface, Turkish language support, a Settings dialog and many smaller improvements. See [CHANGELOG.md](CHANGELOG.md) for details.

## License

MIT License. See [LICENSE](LICENSE) for details.
