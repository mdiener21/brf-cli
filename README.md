# brf

`brf` is a tiny personal Go CLI for quick repo cleanup and later brief repo overviews.
The cleanup commands are implemented in Go and call `git` directly; there is no dependency on `git-cleanup.sh` anymore.

## First commands

The first slice wraps the existing `git-cleanup.sh` commands:

```bash
brf git cleanup merged-branches
brf git cleanup remove-branches
brf git cleanup worktrees
brf git cleanup worktree
brf git cleanup wk
brf git cleanup -wk
brf git cleanup merged-worktrees
brf git cleanup remove-worktrees
brf git cleanup prune
```

## Build

```bash
cd tools/cli/brf
go build -o brf .
```

## Run

```bash
cd tools/cli/brf
go run . help
```

## Configuration

```bash
MAIN_BRANCH=main
```

If `MAIN_BRANCH` is unset, `brf` defaults to `main`.