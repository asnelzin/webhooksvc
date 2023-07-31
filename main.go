package main

import (
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Task represents a task to be executed.
type Task struct {
	ID      string
	Command string
}

// Config represents a configuration file with tasks.
type Config struct {
	Tasks []Task
}

// LoadTasksConfig loads tasks from a YAML file.
func LoadTasksConfig(path string) (map[string]Task, error) {
	doc, err := os.ReadFile(path) // nolint
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(doc, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	tasks := make(map[string]Task)
	for _, task := range cfg.Tasks {
		tasks[task.ID] = task
	}
	return tasks, nil
}

func main() {
	fs := flag.NewFlagSet("webhooksvc", flag.ContinueOnError)
	var (
		listen      = fs.String("listen", ":8080", "server listen address")
		authKey     = fs.String("auth-key", "", "server auth key")
		tasksConfig = fs.String("config", "tasks.yaml", "path to tasks config file")
		timeout     = fs.Duration("timeout", 5*time.Minute, "task execution timeout")
	)
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVars()); err != nil {
		log.Fatalf("[ERROR] could not parse configuration parameters: %v", err)
	}
	if *authKey == "" {
		log.Fatalf("[ERROR] auth key is required")
	}

	log.SetOutput(os.Stdout)

	tasks, err := LoadTasksConfig(*tasksConfig)
	if err != nil {
		log.Fatalf("[ERROR] could not load tasks config: %v", err)
	}

	s := &Server{AuthKey: *authKey, Tasks: tasks, Executor: &ShellExecutor{Timeout: *timeout}}
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal received")
		if serr := s.Shutdown(); serr != nil {
			log.Fatalf("[ERROR] could not shutdown server: %v", err)
		}
	}()

	if err = s.Run(*listen); err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("[ERROR] server failed: %v", err)
		}
	}
}
