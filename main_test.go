package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseWorktreesPorcelain(t *testing.T) {
	output := "worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /repo/feature\nHEAD def456\nbranch refs/heads/feature\n"

	worktrees := parseWorktreesPorcelain(output)
	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}

	if worktrees[0].path != "/repo" || worktrees[0].branch != "main" {
		t.Fatalf("unexpected first worktree: %#v", worktrees[0])
	}

	if worktrees[1].path != "/repo/feature" || worktrees[1].branch != "feature" {
		t.Fatalf("unexpected second worktree: %#v", worktrees[1])
	}
}

func TestSortedCleanupCommands(t *testing.T) {
	commands := sortedCleanupCommands()
	if len(commands) != 6 {
		t.Fatalf("expected 6 commands, got %d", len(commands))
	}

	for i := 1; i < len(commands); i++ {
		if commands[i-1].name > commands[i].name {
			t.Fatalf("commands not sorted: %q before %q", commands[i-1].name, commands[i].name)
		}
	}
}

func TestGitHelpShowsFlattenedCommands(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"git", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected help to succeed, got %v", err)
	}

	help := out.String()
	if !strings.Contains(help, "brf git [command]") {
		t.Fatalf("expected git command usage, got:\n%s", help)
	}
	if !strings.Contains(help, "worktrees") {
		t.Fatalf("expected worktrees command in help, got:\n%s", help)
	}
	if strings.Contains(help, strings.Join([]string{"git", "cleanup"}, " ")) {
		t.Fatalf("expected help not to include old command group, got:\n%s", help)
	}
}

func TestNormalizeArgs(t *testing.T) {
	args := []string{"brf", "git", "-wk"}
	normalized := normalizeArgs(args)
	if normalized[2] != "worktrees" {
		t.Fatalf("expected shorthand to normalize to worktrees, got %q", normalized[2])
	}

	args = []string{"brf", "git", "-wkt"}
	normalized = normalizeArgs(args)
	if normalized[2] != "worktrees" {
		t.Fatalf("expected shorthand to normalize to worktrees, got %q", normalized[2])
	}

	args = []string{"brf", "g", "wkt"}
	normalized = normalizeArgs(args)
	if normalized[1] != "git" || normalized[2] != "worktrees" {
		t.Fatalf("expected g wkt to normalize to git worktrees, got %#v", normalized)
	}

	args = []string{"brf", "git", "-mb"}
	normalized = normalizeArgs(args)
	if normalized[2] != "merged-branches" {
		t.Fatalf("expected shorthand to normalize to merged-branches, got %q", normalized[2])
	}

	args = []string{"brf", "git", "-mbr"}
	normalized = normalizeArgs(args)
	if normalized[2] != "merged-branches" {
		t.Fatalf("expected shorthand to normalize to merged-branches, got %q", normalized[2])
	}

	args = []string{"brf", "git", "-rb"}
	normalized = normalizeArgs(args)
	if normalized[2] != "remove-branches" {
		t.Fatalf("expected shorthand to normalize to remove-branches, got %q", normalized[2])
	}

	args = []string{"brf", "git", "-rbr"}
	normalized = normalizeArgs(args)
	if normalized[2] != "remove-branches" {
		t.Fatalf("expected shorthand to normalize to remove-branches, got %q", normalized[2])
	}

	args = []string{"brf", "git", "-mwk"}
	normalized = normalizeArgs(args)
	if normalized[2] != "merged-worktrees" {
		t.Fatalf("expected shorthand to normalize to merged-worktrees, got %q", normalized[2])
	}

	args = []string{"brf", "git", "-mwkt"}
	normalized = normalizeArgs(args)
	if normalized[2] != "merged-worktrees" {
		t.Fatalf("expected shorthand to normalize to merged-worktrees, got %q", normalized[2])
	}

	args = []string{"brf", "git", "-rwk"}
	normalized = normalizeArgs(args)
	if normalized[2] != "remove-worktrees" {
		t.Fatalf("expected shorthand to normalize to remove-worktrees, got %q", normalized[2])
	}

	args = []string{"brf", "git", "-rwkt"}
	normalized = normalizeArgs(args)
	if normalized[2] != "remove-worktrees" {
		t.Fatalf("expected shorthand to normalize to remove-worktrees, got %q", normalized[2])
	}

	nonShortcut := []string{"brf", "git", "worktree"}
	normalized = normalizeArgs(nonShortcut)
	if normalized[2] != "worktree" {
		t.Fatalf("expected non-shortcut argument to remain unchanged, got %q", normalized[2])
	}
}
