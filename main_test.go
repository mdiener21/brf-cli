package main

import "testing"

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

func TestNormalizeArgs(t *testing.T) {
	args := []string{"brf", "git", "cleanup", "-wk"}
	normalized := normalizeArgs(args)
	if normalized[3] != "worktrees" {
		t.Fatalf("expected shorthand to normalize to worktrees, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-wkt"}
	normalized = normalizeArgs(args)
	if normalized[3] != "worktrees" {
		t.Fatalf("expected shorthand to normalize to worktrees, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-mb"}
	normalized = normalizeArgs(args)
	if normalized[3] != "merged-branches" {
		t.Fatalf("expected shorthand to normalize to merged-branches, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-mbr"}
	normalized = normalizeArgs(args)
	if normalized[3] != "merged-branches" {
		t.Fatalf("expected shorthand to normalize to merged-branches, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-rb"}
	normalized = normalizeArgs(args)
	if normalized[3] != "remove-branches" {
		t.Fatalf("expected shorthand to normalize to remove-branches, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-rbr"}
	normalized = normalizeArgs(args)
	if normalized[3] != "remove-branches" {
		t.Fatalf("expected shorthand to normalize to remove-branches, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-mwk"}
	normalized = normalizeArgs(args)
	if normalized[3] != "merged-worktrees" {
		t.Fatalf("expected shorthand to normalize to merged-worktrees, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-mwkt"}
	normalized = normalizeArgs(args)
	if normalized[3] != "merged-worktrees" {
		t.Fatalf("expected shorthand to normalize to merged-worktrees, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-rwk"}
	normalized = normalizeArgs(args)
	if normalized[3] != "remove-worktrees" {
		t.Fatalf("expected shorthand to normalize to remove-worktrees, got %q", normalized[3])
	}

	args = []string{"brf", "git", "cleanup", "-rwkt"}
	normalized = normalizeArgs(args)
	if normalized[3] != "remove-worktrees" {
		t.Fatalf("expected shorthand to normalize to remove-worktrees, got %q", normalized[3])
	}

	nonShortcut := []string{"brf", "git", "cleanup", "worktree"}
	normalized = normalizeArgs(nonShortcut)
	if normalized[3] != "worktree" {
		t.Fatalf("expected non-shortcut argument to remain unchanged, got %q", normalized[3])
	}
}
