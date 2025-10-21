# nt

A simple README describing what is used to code this project (Go) and clear installation and usage instructions for both Windows and Ubuntu.

## What this project uses

- Language: Go (100%)
- Minimum Go version: 1.20 (recommended latest stable)
- Go modules (go.mod)
- Standard library and any modules listed in go.mod (if present)

## Prerequisites

- Git (to clone the repo)
- Go toolchain installed (see OS-specific instructions below)

## Install and build (Ubuntu)

1. Install Go (recommended: download from https://go.dev/dl or use snap):

   Option A — official tarball (recommended for latest Go):

   - Download the tarball for Linux from https://go.dev/dl
   - Extract and install (example for Go 1.20+):
     sudo tar -C /usr/local -xzf go1.20.linux-amd64.tar.gz
   - Add /usr/local/go/bin to your PATH (add to ~/.profile or ~/.bashrc):
     export PATH=$PATH:/usr/local/go/bin
   - Reload your profile: source ~/.profile

   Option B — apt (may provide older Go):

   sudo apt update
   sudo apt install -y golang-go

2. Clone the repository and build:

   git clone https://github.com/faeiiz/nt.git
   cd nt

   # Use modules (recommended)
   go mod download

   # Build
   go build -o nt

   # Run
   ./nt

3. Install to your Go bin (optional):

   # Installs to $(go env GOPATH)/bin or GOBIN if set
   go install ./...

   # Make sure GOPATH/bin or GOBIN is in your PATH
   export PATH=$PATH:$(go env GOPATH)/bin

## Install and build (Windows)

1. Install Go:

   - Download the Windows installer (MSI) from https://go.dev/dl and run it.
   - The installer updates your PATH for you (if you accept defaults).

2. Clone the repository and build (PowerShell):

   git clone https://github.com/faeiiz/nt.git
   cd nt

   # Use modules (recommended)
   go mod download

   # Build
   go build -o nt.exe

   # Run
   .\nt.exe

3. Install to your Go bin (optional):

   go install ./...

   # Ensure %GOPATH%\bin or GOBIN is in your PATH

## Use (examples)

- Build and run directly:
  go build -o nt && ./nt

- Run without building (for development):
  go run ./...

- Install for global usage:
  go install ./...
  # then run `nt` from a terminal (if GOPATH/bin is on PATH)

## Environment / Modules

- If the repository uses Go modules (has go.mod), the build commands above will use modules.
- If not, set GOPATH and use the traditional GOPATH workflow.

## Troubleshooting

- "go: command not found": ensure Go is installed and on PATH.
- "permission denied" when installing to /usr/local: use sudo for installation steps or install Go to a user-writable directory and update PATH.
- If builds fail due to missing dependencies, run `go mod tidy` and `go mod download`.

## Contributing

Contributions are welcome. Open issues or pull requests with fixes or improvements.

## License

If there is a LICENSE file in the repository, this project follows that license. If not, add a LICENSE file (for example, MIT) to make the license explicit.


(This README was generated and added by GitHub Copilot.)