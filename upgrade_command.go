package main

import (
	"context"
	"fmt"

	"brf/pkg/updater"
	"github.com/spf13/cobra"
)

const repository = "mdiener21/brf-cli"

type upgradeFunc func(context.Context, string) (updater.Result, error)

func newUpgradeCmd(currentVersion string, run upgradeFunc) *cobra.Command {
	return &cobra.Command{
		Use:     "upgrade",
		Aliases: []string{"update"},
		Short:   "Upgrade brf to the latest release",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeUpgrade(cmd, currentVersion, run)
		},
	}
}

func executeUpgrade(cmd *cobra.Command, currentVersion string, run upgradeFunc) error {
	result, err := run(cmd.Context(), currentVersion)
	if err != nil {
		return fmt.Errorf("upgrade brf: %w", err)
	}

	if result.AlreadyCurrent {
		fmt.Fprintf(cmd.OutOrStdout(), "brf %s is already the latest release.\n", currentVersion)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Upgraded brf from %s to %s.\n", currentVersion, result.Version)
	return nil
}

func runUpgrade(ctx context.Context, currentVersion string) (updater.Result, error) {
	return updater.New(repository).Upgrade(ctx, currentVersion)
}
