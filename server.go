package main

import (
	"context"
	"crypto/subtle"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Executor interface {
	Exec(ctx context.Context, command string, out io.Writer) error
}

type Server struct {
	AuthKey  string
	Tasks    map[string]Task
	Executor Executor

	runLock sync.Mutex
	httpsrv *http.Server
}

func (s *Server) Run(addr string) error {
	s.runLock.Lock()
	s.httpsrv = &http.Server{
		Addr:    addr,
		Handler: s.routes(),
	}
	s.runLock.Unlock()
	return s.httpsrv.ListenAndServe()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	s.runLock.Lock()
	err := s.httpsrv.Shutdown(ctx)
	s.runLock.Unlock()
	return err
}

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("App", "webhooksvc@v1.0.0"))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(s.auth)

	r.Post("/tasks/{taskID}/execute", s.executeTaskCtrl)
	return r
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte(s.AuthKey)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type ExecLogWriter struct {
	Logger *log.Logger
}

func (l *ExecLogWriter) Write(bytes []byte) (n int, err error) {
	l.Logger.Printf("[INFO] > %s", bytes)
	return len(bytes), nil
}

func (s *Server) executeTaskCtrl(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	task, ok := s.Tasks[taskID]
	if !ok {
		http.Error(w, "task not found", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] exec task %s", task.ID)
	err := s.Executor.Exec(r.Context(), task.Command, &ExecLogWriter{Logger: log.Default()})
	if err != nil {
		http.Error(w, "could not execute task", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(`{"status": "ok", "task": "` + task.ID + `"}`))
}
