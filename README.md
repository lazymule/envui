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

## Features

- **Visual Editor** - Table view of all variables, click to edit inline
- **Secret Masking** - Auto-detect and mask sensitive values (API_KEY, PASSWORD, TOKEN, etc.)
- **Multi-file Support** - Switch between `.env`, `.env.local`, `.env.staging`, etc.
- **Source Tracking** - See which variables come from file vs shell environment
- **Comment Preservation** - Your comments stay intact when saving

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
