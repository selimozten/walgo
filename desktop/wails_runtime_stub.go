//go:build !desktop

package main

import "context"

func windowMinimise(_ context.Context)           {}
func windowToggleMaximise(_ context.Context)     {}
func appQuit(_ context.Context)                  {}
func browserOpenURL(_ context.Context, _ string) {}

func openDirectoryDialog(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
