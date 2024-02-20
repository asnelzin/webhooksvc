# Webhook Service

## Overview

This Webhook Service enables the execution of predefined tasks identified by a task ID, secured with an authentication key.

## Configuration

The service supports configuration through both environment variables and command-line flags:

- **`listen`**: Server listen address. Defaults to `:8080`.
- **`auth-key`**: Server authentication key, required for validating requests.
- **`config`**: Path to the tasks configuration file. Defaults to `tasks.yaml`.
- **`timeout`**: Task execution timeout. Defaults to `5m` (5 minutes).

Tasks are defined in a YAML configuration file (`tasks.yaml`), with each task specifying an ID and command:

```yaml
tasks:
  - id: test
    command: |
      echo "Hello World"
```

### Running the Service

Use Docker Compose to run the service, which specifies default environment variables and mounts the tasks configuration file as a volume for easy customization:

```yaml
services:
  webhooksvc:
    build: .
    image: ghcr.io/asnelzin/webhooksvc:latest
    container_name: "webhooksvc"
    ports:
      - "80:8080"
    environment:
      - AUTH_KEY=secret
      - CONFIG=/srv/etc/tasks.yml
    volumes:
      - ./tasks.yml:/srv/etc/tasks.yml:ro
```

This configuration:
- Sets the `AUTH_KEY` environment variable to your chosen secret, which is required for task execution.
- Maps the `CONFIG` environment variable to the location of the `tasks.yml` file inside the container.
- Mounts the `tasks.yml` file from your host to `/srv/etc/tasks.yml` inside the container, allowing the service to read your tasks configuration.

To start the service:

```bash
docker-compose up -d
```

This command exposes the service on port 80 of your host, making it accessible for task execution requests.

### Usage

To execute a task, send a POST request to `/tasks/{taskID}/execute`, with `{taskID}` being the identifier of the desired task. The request must include the authentication key in the `Authorization` header.

Example request using `curl`:

```bash
curl -X POST http://localhost/tasks/test/execute      -H "Authorization: secret"
```

This request triggers the execution of the task associated with the `test` ID, given that the correct authentication key is provided.
