package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer_Run(t *testing.T) {
	srv := Server{}

	errchan := make(chan error, 1)
	go func() {
		err := srv.Run("localhost:55055")
		errchan <- err
	}()
	time.Sleep(1 * time.Second)

	err := srv.Shutdown()
	require.NoError(t, err)

	err = <-errchan
	require.Error(t, err)
	assert.Equal(t, http.ErrServerClosed, err)
}

func TestServer_auth(t *testing.T) {
	srv := Server{AuthKey: "blah"}
	okCtrl := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(srv.auth(okCtrl))
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "blah")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest("POST", ts.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "wrong-key")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestServer_ping(t *testing.T) {
	srv := Server{AuthKey: "blah"}
	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ping")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

type MockExecutor struct {
	ExecFn func(command string, out io.Writer) error
}

func (m *MockExecutor) Exec(ctx context.Context, command string, out io.Writer) error {
	return m.ExecFn(command, out)
}

func TestServer_executeTaskCtrl(t *testing.T) {
	srv := Server{
		AuthKey: "blah",
		Tasks: map[string]Task{
			"task1": {ID: "task1", Command: "echo 'hello'"},
		},
		Executor: &MockExecutor{
			ExecFn: func(command string, out io.Writer) error {
				assert.Equal(t, "echo 'hello'", command)
				return nil
			},
		},
	}
	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL+"/tasks/task1/execute", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "blah")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"status": "ok", "task": "task1"}`, string(body))
}
