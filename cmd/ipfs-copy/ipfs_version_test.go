package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasShellStreamListSupport(t *testing.T) {
	require.False(t, hasShellStreamPinListSupport("0.4.22"))
	require.True(t, hasShellStreamPinListSupport("0.5.0"))
	require.True(t, hasShellStreamPinListSupport("0.5.1"))
	require.True(t, hasShellStreamPinListSupport("0.7.0"))
	require.True(t, hasShellStreamPinListSupport("0.9.1"))
}
