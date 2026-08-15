# brf

`brf` is a tiny personal Go CLI for quick repo cleanup and later brief repo overviews.
The cleanup commands are implemented in Go and call `git` directly; there is no dependency on `git-cleanup.sh` anymore.

## First commands

The first slice wraps the existing `git-cleanup.sh` commands:

```bash
brf git cleanup merged-branches
brf git cleanup -mb
brf git cleanup remove-branches
brf git cleanup -rb
brf git cleanup worktrees
brf git cleanup worktree
brf git cleanup wk
brf git cleanup -wk
brf git cleanup -wkt
brf git cleanup merged-worktrees
brf git cleanup remove-worktrees
brf git cleanup prune
```

## Help

Every command supports help flags from Cobra:

```bash
brf --help
brf -h
brf git cleanup --help
brf git cleanup -h
```

## Build Locally (from source)

From the repository root:

```bash
go build -o brf .
./brf --version
```

## Install Locally (from source)

Install to your Go bin directory:

```bash
go install .
brf --version
```

Optional custom install directory:

```bash
GOBIN="$HOME/.local/bin" go install .
"$HOME/.local/bin/brf" --version
```

## Install Prebuilt CLI from GitHub Releases

### Linux (amd64)

```bash
TAG="2026-08-15-1435-ga5kh2b"
curl -fLO "https://github.com/mdiener21/brf-cli/releases/download/${TAG}/brf_${TAG}_linux_amd64.tar.gz"
tar -xzf "brf_${TAG}_linux_amd64.tar.gz"
install -m 0755 "brf_${TAG}_linux_amd64/brf" "$HOME/.local/bin/brf"
"$HOME/.local/bin/brf" --version
```

### Linux (arm64)

```bash
TAG="2026-08-15-1435-ga5kh2b"
curl -fLO "https://github.com/mdiener21/brf-cli/releases/download/${TAG}/brf_${TAG}_linux_arm64.tar.gz"
tar -xzf "brf_${TAG}_linux_arm64.tar.gz"
install -m 0755 "brf_${TAG}_linux_arm64/brf" "$HOME/.local/bin/brf"
"$HOME/.local/bin/brf" --version
```

### Windows (amd64)

Download and extract:

```text
brf_<TAG>_windows_amd64.zip
```

Then place `brf.exe` somewhere on your `PATH` and run:

```powershell
brf.exe --version
```

### Verify Release Checksums (optional)

```bash
TAG="2026-08-15-1435-ga5kh2b"
curl -fLO "https://github.com/mdiener21/brf-cli/releases/download/${TAG}/brf_${TAG}_checksums.txt"
sha256sum -c "brf_${TAG}_checksums.txt" --ignore-missing
```

## Configuration

```bash
MAIN_BRANCH=main
```

If `MAIN_BRANCH` is unset, `brf` defaults to `main`.