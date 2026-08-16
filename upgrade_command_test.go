package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"brf/pkg/updater"
)

func TestUpgradeCommandAliases(t *testing.T) {
	for _, name := range []string{"upgrade", "update"} {
		t.Run(name, func(t *testing.T) {
			called := false
			cmd := newRootCmdWithUpgrade(func(context.Context, string) (updater.Result, error) {
				called = true
				return updater.Result{Version: "0.6.0"}, nil
			})
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{name})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("expected %s to succeed, got %v", name, err)
			}
			if !called {
				t.Fatalf("expected %s to invoke updater", name)
			}
		})
	}
}

func TestUpgradeFlag(t *testing.T) {
	called := false
	cmd := newRootCmdWithUpgrade(func(context.Context, string) (updater.Result, error) {
		called = true
		return updater.Result{Version: "0.6.0"}, nil
	})
	cmd.SetArgs([]string{"--upgrade"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected --upgrade to succeed, got %v", err)
	}
	if !called {
		t.Fatal("expected --upgrade to invoke updater")
	}
}

func TestExecuteUpgrade(t *testing.T) {
	cmd := newUpgradeCmd("0.5.0", func(context.Context, string) (updater.Result, error) {
		return updater.Result{Version: "0.6.0"}, nil
	})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected upgrade to succeed, got %v", err)
	}
	if got := out.String(); !strings.Contains(got, "Upgraded brf from 0.5.0 to 0.6.0") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestExecuteUpgradeAlreadyCurrent(t *testing.T) {
	cmd := newUpgradeCmd("0.5.0", func(context.Context, string) (updater.Result, error) {
		return updater.Result{Version: "v0.5.0", AlreadyCurrent: true}, nil
	})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected upgrade to succeed, got %v", err)
	}
	if got := out.String(); !strings.Contains(got, "already the latest release") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestExecuteUpgradeError(t *testing.T) {
	cmd := newUpgradeCmd("0.5.0", func(context.Context, string) (updater.Result, error) {
		return updater.Result{}, errors.New("network unavailable")
	})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "upgrade brf: network unavailable") {
		t.Fatalf("expected actionable error, got %v", err)
	}
}
