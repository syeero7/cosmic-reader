# Cosmic Reader

A cross-platform desktop app for reading and managing digital comics. Supports CBZ, CBR, CBA, CB7 and CBT archive formats.

![screenshot](./screenshot.webp)

## Getting Started

Download [latest release](https://github.com/syeero7/cosmic-reader/releases/latest)

```bash
# add execute permission
chmod +x ./cosmic-reader-linux-amd64

./cosmic-reader-linux-amd64
# or
./cosmic-reader-linux-amd64  /path/to/comic.cbz
```

### Build from Source

Install wails and required dependencies. [Wails docs](https://wails.io/docs/gettingstarted/installation/)

Clone the repository

```bash
git clone https://github.com/syeero7/cosmic-reader
cd cosmic-reader
```

Compile the binary

```bash
# linux
wails build -upx -trimpath -platform=linux

# windows
wails build -upx -trimpath -platform=windows
```

## Local Development

Install wails and required dependencies. [Wails docs](https://wails.io/docs/gettingstarted/installation/)

```bash
# Install dependencies
go mod tidy
npm install --prefix ./frontend

# Initialize go workspace
go work init .

# Add local Wails module to the workspace
go work use /home/<username>/go/pkg/mod/github.com/wailsapp/wails/<wails_version>

wails dev
```
