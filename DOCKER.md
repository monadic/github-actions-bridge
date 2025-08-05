# Docker Guide for GitHub Actions Bridge

This guide covers running the GitHub Actions Bridge using Docker and Docker Compose.

## Quick Start with Docker Compose

### 1. Setup Environment

```bash
# Copy the example environment file
cp .env.example .env

# Edit with your ConfigHub credentials
vim .env
```

### 2. Start the Bridge

```bash
# Start in detached mode
docker-compose up -d

# Or run in foreground to see logs
docker-compose up
```

### 3. Verify Health

```bash
# Check health endpoint
curl http://localhost:8080/health

# View logs
docker-compose logs -f actions-bridge

# Check container status
docker-compose ps
```

### 4. Stop the Bridge

```bash
# Stop and remove containers
docker-compose down

# Stop and remove containers + volumes
docker-compose down -v
```

## Docker Image

### Building the Image

```bash
# Build using docker-compose
docker-compose build

# Or build directly
docker build -t confighub/actions-bridge:latest .
```

### Running Standalone

```bash
docker run -d \
  --name actions-bridge \
  -e CONFIGHUB_WORKER_ID=your-worker-id \
  -e CONFIGHUB_WORKER_SECRET=your-worker-secret \
  -e CONFIGHUB_URL=https://api.confighub.com \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v actions-bridge-data:/var/lib/actions-bridge \
  -p 8080:8080 \
  --restart unless-stopped \
  confighub/actions-bridge:latest
```

## Configuration

### Environment Variables

All configuration is done through environment variables. See `.env.example` for available options.

### Docker Socket Access

The bridge requires access to the Docker socket to run workflows with act. The compose file mounts `/var/run/docker.sock`.

**Security Note**: This gives the container full access to Docker. Only run trusted images.

### Volumes

- `actions-bridge-data`: Persistent storage for workspaces and artifacts
- `/var/run/docker.sock`: Docker socket for running containers

### Networking

The bridge exposes port 8080 for health checks and metrics. In production, you may want to:

1. Use a reverse proxy (nginx, traefik)
2. Enable TLS/SSL
3. Restrict access to health endpoints

## Monitoring with Prometheus

### Enable Monitoring

```bash
# Start with monitoring profile
docker-compose --profile monitoring up -d

# Access Prometheus
open http://localhost:9090
```

### Available Metrics

- Workflow execution count
- Execution duration
- Success/failure rates
- Concurrent executions
- Resource usage

## Production Considerations

### Security

1. **User Permissions**: The container runs as non-root user (uid 1000)
2. **Read-Only Root**: Consider adding `read_only: true` with tmpfs mounts
3. **Resource Limits**: Add CPU/memory limits in production

```yaml
services:
  actions-bridge:
    # ... other config ...
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
```

### High Availability

For production deployments:

1. Use external volume drivers (NFS, EBS, etc.)
2. Run multiple instances with different WORKER_IDs
3. Use a load balancer for health checks
4. Enable distributed tracing

### Logging

```yaml
services:
  actions-bridge:
    # ... other config ...
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose logs actions-bridge

# Verify Docker socket access
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock alpine docker version
```

### Permission Denied

```bash
# Fix Docker socket permissions
sudo chmod 666 /var/run/docker.sock

# Or add user to docker group
sudo usermod -aG docker $USER
```

### Workflow Execution Fails

```bash
# Check act is working
docker exec actions-bridge act --version

# Test workflow directly
docker exec actions-bridge act -W /test/fixtures/workflows/simple.yml
```

### Clean Restart

```bash
# Stop everything
docker-compose down

# Remove volumes
docker volume rm github-actions-bridge_actions-bridge-data

# Rebuild and start
docker-compose build --no-cache
docker-compose up -d
```

## Development with Docker

### Hot Reload

For development, mount source code:

```yaml
services:
  actions-bridge:
    volumes:
      - .:/app
      - /app/bin  # Exclude binaries
    command: go run cmd/actions-bridge/main.go
```

### Debug Mode

```yaml
services:
  actions-bridge:
    environment:
      DEBUG: "true"
    command: ["dlv", "debug", "--headless", "--listen=:2345", "--api-version=2", "cmd/actions-bridge/main.go"]
    ports:
      - "2345:2345"  # Delve debugger
```