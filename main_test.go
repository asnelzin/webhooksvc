package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadTasksConfig(t *testing.T) {
	tasks, err := LoadTasksConfig("testdata/config_example.yml")
	assert.NoError(t, err)

	assert.Equal(t, 2, len(tasks))
	assert.Equal(t, Task{ID: "hello-world", Command: "echo \"Hello World\"\n"}, tasks["hello-world"])
	assert.Equal(t, Task{ID: "restart-compose", Command: "docker-compose down\ndocker-compose up -d\n"}, tasks["restart-compose"])

	_, err = LoadTasksConfig("unexpected_file.yml")
	assert.Error(t, err)
}
