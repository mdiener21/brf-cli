package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const version = "0.5.0"

type cleanupCommand struct {
	name        string
	description string
}

var cleanupCommands = []cleanupCommand{
	{name: "merged-branches", description: "List local branches fully merged into the main branch"},
	{name: "remove-branches", description: "Delete local branches fully merged into the main branch"},
	{name: "worktrees", description: "Show worktrees and merge status"},
	{name: "merged-worktrees", description: "List worktrees fully merged into the main branch"},
	{name: "remove-worktrees", description: "Remove worktrees fully merged into the main branch"},
	{name: "prune", description: "Prune stale Git worktree metadata"},
}

func main() {
	os.Args = normalizeArgs(os.Args)
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func normalizeArgs(args []string) []string {
	if len(args) < 3 {
		return args
	}

	if args[1] == "g" {
		normalized := append([]string(nil), args...)
		normalized[1] = "git"
		args = normalized
	}

	if args[1] == "git" {
		shorthands := map[string]string{
			"-wk":   "worktrees",
			"-wkt":  "worktrees",
			"wkt":   "worktrees",
			"-mb":   "merged-branches",
			"-mbr":  "merged-branches",
			"-rb":   "remove-branches",
			"-rbr":  "remove-branches",
			"-mwk":  "merged-worktrees",
			"-mwkt": "merged-worktrees",
			"-rwk":  "remove-worktrees",
			"-rwkt": "remove-worktrees",
		}

		if expanded, ok := shorthands[args[2]]; ok {
			normalized := append([]string(nil), args...)
			normalized[2] = expanded
			return normalized
		}
	}

	return args
}

func newRootCmd() *cobra.Command {
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:           "brf",
		Short:         "Small Git cleanup helper",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Fprintln(cmd.OutOrStdout(), version)
				return nil
			}
			usage(cmd.OutOrStdout())
			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version information")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	})
	rootCmd.AddCommand(newGitCmd())

	return rootCmd
}

func newGitCmd() *cobra.Command {
	gitCmd := &cobra.Command{
		Use:     "git",
		Aliases: []string{"g"},
		Short:   "Git utilities",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "Usage: brf git <command>")
		},
	}

	gitCmd.AddCommand(newCleanupLeafCmd(
		"merged-branches",
		"List local branches fully merged into the main branch",
		func(mainBranch string) error {
			return listMergedBranches(os.Stdout, mainBranch)
		},
	))
	gitCmd.AddCommand(newCleanupLeafCmd(
		"remove-branches",
		"Delete local branches fully merged into the main branch",
		func(mainBranch string) error {
			return removeMergedBranches(mainBranch)
		},
	))
	gitCmd.AddCommand(newCleanupLeafCmd(
		"worktrees",
		"List and show worktrees and merge status",
		func(mainBranch string) error {
			return printWorktreeStatus(os.Stdout, mainBranch)
		},
		"worktree",
		"wk",
		"wkt",
	))
	gitCmd.AddCommand(newCleanupLeafCmd(
		"merged-worktrees",
		"List worktrees fully merged into the main branch",
		func(mainBranch string) error {
			return printMergedWorktrees(os.Stdout, mainBranch)
		},
	))
	gitCmd.AddCommand(newCleanupLeafCmd(
		"remove-worktrees",
		"Remove worktrees fully merged into the main branch",
		func(mainBranch string) error {
			return removeMergedWorktrees(mainBranch)
		},
	))
	gitCmd.AddCommand(newCleanupLeafCmd(
		"prune",
		"Prune stale Git worktree metadata",
		func(mainBranch string) error {
			_ = mainBranch
			return runGit("worktree", "prune")
		},
	))

	return gitCmd
}

func newCleanupLeafCmd(use, short string, runner func(mainBranch string) error, aliases ...string) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Short:   short,
		Aliases: aliases,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mainBranch := mainBranchName()
			if err := checkRepo(mainBranch); err != nil {
				return err
			}

			err := runner(mainBranch)
			if err == nil {
				return nil
			}

			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}

			return fmt.Errorf("brf git: %v", err)
		},
	}
}

func usage(w io.Writer) {
	fmt.Fprintf(w, "brf %s\n\n", version)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  brf git <command>")
	fmt.Fprintln(w, "  brf g <command>")
	fmt.Fprintln(w, "  brf help")
	fmt.Fprintln(w, "  brf --version")
	fmt.Fprintln(w, "\nCommands:")
	for _, command := range sortedCleanupCommands() {
		fmt.Fprintf(w, "  %-18s %s\n", command.name, command.description)
	}
}
func sortedCleanupCommands() []cleanupCommand {
	commands := append([]cleanupCommand(nil), cleanupCommands...)
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].name < commands[j].name
	})
	return commands
}

func isCleanupCommand(name string) bool {
	for _, command := range cleanupCommands {
		if command.name == name {
			return true
		}
	}
	return false
}

func mainBranchName() string {
	if branch := os.Getenv("MAIN_BRANCH"); branch != "" {
		return branch
	}

	return "main"
}

func checkRepo(mainBranch string) error {
	if _, err := runGitOutput("rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("brf: not inside a Git repository")
	}

	if _, err := runGitOutput("rev-parse", "--verify", mainBranch); err != nil {
		return fmt.Errorf("brf: branch %q does not exist", mainBranch)
	}

	return nil
}

func listMergedBranches(w io.Writer, mainBranch string) error {
	output, err := runGitOutput("for-each-ref", "--format=%(refname:short)", "--merged="+mainBranch, "refs/heads")
	if err != nil {
		return err
	}

	for _, branch := range splitLines(output) {
		if branch == "" || branch == mainBranch {
			continue
		}
		fmt.Fprintln(w, branch)
	}

	return nil
}

func removeMergedBranches(mainBranch string) error {
	output, err := runGitOutput("for-each-ref", "--format=%(refname:short)", "--merged="+mainBranch, "refs/heads")
	if err != nil {
		return err
	}

	for _, branch := range splitLines(output) {
		if branch == "" || branch == mainBranch {
			continue
		}

		fmt.Fprintf(os.Stdout, "Deleting merged branch: %s\n", branch)
		if err := runGit("branch", "-d", branch); err != nil {
			return err
		}
	}

	return nil
}

func printWorktreeStatus(w io.Writer, mainBranch string) error {
	for _, worktree := range parseWorktreesPorcelain(mustRunGitOutput("worktree", "list", "--porcelain")) {
		if worktree.branch == "" {
			continue
		}

		branch := worktree.branch
		if branch == mainBranch {
			continue
		}

		merged, err := isAncestor(branch, mainBranch)
		if err != nil {
			return err
		}

		if merged {
			fmt.Fprintf(w, "MERGED      %-40s %s\n", branch, worktree.path)
			continue
		}

		fmt.Fprintf(w, "NOT MERGED  %-40s %s\n", branch, worktree.path)
	}

	return nil
}

func printMergedWorktrees(w io.Writer, mainBranch string) error {
	for _, worktree := range parseWorktreesPorcelain(mustRunGitOutput("worktree", "list", "--porcelain")) {
		if worktree.branch == "" || worktree.branch == mainBranch {
			continue
		}

		merged, err := isAncestor(worktree.branch, mainBranch)
		if err != nil {
			return err
		}
		if !merged {
			continue
		}

		fmt.Fprintf(w, "%-40s %s\n", worktree.branch, worktree.path)
	}

	return nil
}

func removeMergedWorktrees(mainBranch string) error {
	for _, worktree := range parseWorktreesPorcelain(mustRunGitOutput("worktree", "list", "--porcelain")) {
		if worktree.branch == "" || worktree.branch == mainBranch {
			continue
		}

		merged, err := isAncestor(worktree.branch, mainBranch)
		if err != nil {
			return err
		}
		if !merged {
			continue
		}

		fmt.Fprintf(os.Stdout, "Removing merged worktree: %s (%s)\n", worktree.path, worktree.branch)
		if err := runGit("worktree", "remove", worktree.path); err != nil {
			return err
		}
	}

	return runGit("worktree", "prune")
}

type worktreeInfo struct {
	path   string
	branch string
}

func parseWorktreesPorcelain(output string) []worktreeInfo {
	var worktrees []worktreeInfo
	var current worktreeInfo
	var haveCurrent bool

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			if haveCurrent {
				worktrees = append(worktrees, current)
			}
			current = worktreeInfo{path: strings.TrimPrefix(line, "worktree ")}
			haveCurrent = true
		case strings.HasPrefix(line, "branch refs/heads/"):
			current.branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}

	if haveCurrent {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

func splitLines(output string) []string {
	if output == "" {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return lines
}

func isAncestor(branch, mainBranch string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", branch, mainBranch)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
			return false, err
		}
		return false, err
	}

	return true, nil
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runGitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	return string(output), err
}

func mustRunGitOutput(args ...string) string {
	output, err := runGitOutput(args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "brf: %v\n", err)
		os.Exit(1)
	}

	return output
}
