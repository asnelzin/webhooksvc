package main

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestShellExecutor_Exec(t *testing.T) {
	se := &ShellExecutor{Timeout: time.Second}

	buf := bytes.NewBuffer(nil)
	err := se.Exec(context.Background(), "echo hello", buf)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", buf.String())
}

func TestShellExecutor_ExecMultipleLines(t *testing.T) {
	se := &ShellExecutor{Timeout: time.Second}

	buf := bytes.NewBuffer(nil)
	err := se.Exec(context.Background(), "echo hello\necho world", buf)
	require.NoError(t, err)
	assert.Equal(t, "hello\nworld\n", buf.String())
}

func TestShellExecutor_ExecTimeOut(t *testing.T) {
	se := &ShellExecutor{Timeout: 100 * time.Millisecond}

	buf := bytes.NewBuffer(nil)
	st := time.Now()
	err := se.Exec(context.Background(), "sleep 1 && sleep 1", buf)
	require.Error(t, err)
	assert.True(t, time.Since(st) < 2*time.Second)
}
