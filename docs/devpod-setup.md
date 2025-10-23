# DevPod Setup Guide

This guide covers setting up DevPod with vCluster and Authentik OIDC for secure development environments.

## Prerequisites

- DevPod CLI installed
- Access to organization Authentik instance
- vCluster infrastructure access
- kubectl configured (optional, handled by DevPod)

## Initial Setup

### 1. Install DevPod

```bash
# macOS
brew install devpod

# Linux
curl -L -o devpod "https://github.com/loft-sh/devpod/releases/latest/download/devpod-linux-amd64" && sudo install -c -m 0755 devpod /usr/local/bin

# Windows
winget install loft.devpod
```

### 2. Configure Kubernetes Provider

```bash
devpod provider add kubernetes
```

### 3. Set OIDC Authentication (if using Authentik)

If your cluster uses Authentik OIDC authentication, configure these environment variables:

```bash
export KUBERNETES_OIDC_ISSUER_URL="https://authentik.example.com/application/o/devpod/"
export KUBERNETES_OIDC_CLIENT_ID="your-devpod-client-id"
export KUBERNETES_OIDC_CLIENT_SECRET="your-devpod-client-secret"
```

For local development clusters, you may not need OIDC authentication.

These values should be provided by your infrastructure team.

## Creating Your First Workspace

### 1. Clone and Start

```bash
git clone https://github.com/simonbrundin/plan.git
cd plan
devpod up .
```

### 2. Authenticate

If using OIDC authentication, DevPod will automatically redirect you to Authentik for authentication. Complete the OIDC flow.

### 3. Workspace Initialization

DevPod will:
- Create your personal namespace (`devpod-<username>`)
- Deploy PostgreSQL and Hasura in your namespace
- Install dependencies using Bun
- Start Tilt for local development with all services
- Set up port forwarding for database (5432), Hasura (8080), and frontend (3000)

## Development Workflow

### Starting Development

```bash
# The workspace will automatically start services via Tilt
# Access your application at http://localhost:3000
# Hasura console at http://localhost:8080/console
# Database at localhost:5432
```

### Understanding the Setup

The DevPod environment includes:

- **PostgreSQL**: Database server running in Kubernetes
- **Hasura**: GraphQL API layer with console access
- **Frontend**: Nuxt.js development server with hot reloading
- **Tilt**: Manages all services and provides port forwarding

All services run in your personal Kubernetes namespace for isolation.

### Running Tests

```bash
cd frontend
bun run test
```

### Database Operations

```bash
# Push schema changes
bun run db:push

# Open Drizzle Studio
bun run db:studio
```

### IDE Integration

DevPod supports VS Code and other IDEs. The workspace will open automatically with your configured IDE.

## Troubleshooting

### Authentication Issues

- Verify OIDC environment variables are set correctly
- Check Authentik configuration with your infrastructure team
- Ensure your user has the correct permissions in Authentik

### vCluster Connection Issues

- Verify network connectivity to the vCluster
- Check namespace creation permissions
- Contact infrastructure team for vCluster access issues

### Service Startup Issues

- Check Tilt logs: `tilt logs` in the workspace
- Verify resource quotas in your namespace
- Check pod status: `kubectl get pods -n devpod-<username>`
- Ensure all required environment variables are set
- Check that the Kubernetes cluster is accessible

### Database Connection Issues

- Verify PostgreSQL pod is running: `kubectl get pods -n devpod-<username>`
- Check database logs: `kubectl logs -n devpod-<username> deployment/plan-dev-postgres`
- Ensure Hasura can connect to the database

### Frontend Development Issues

- Check that all dependencies are installed: `cd frontend && bun install`
- Verify the development server is running: `ps aux | grep "bun run dev"`
- Check browser console for any JavaScript errors

## Security Features

- **Namespace Isolation**: Each developer gets isolated namespaces
- **OIDC Authentication**: Secure authentication through Authentik
- **Network Policies**: Restricted network access between namespaces
- **Resource Limits**: Automatic CPU and memory limits per workspace

## Advanced Configuration

### Custom Resource Limits

Edit `.devpod.yaml` to modify resource limits:

```yaml
workspace:
  devContainer:
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "2"
        memory: "4Gi"
```

### Additional IDE Extensions

Add extensions to `.devpod.yaml`:

```yaml
ide:
  options:
    extensions:
      - ms-python.python
      - github.copilot
```

## Infrastructure Team Tasks

This section is for infrastructure teams setting up DevPod:

### Authentik Configuration

1. Create OIDC provider in Authentik
2. Configure client credentials
3. Set up user groups and permissions

### vCluster Setup

1. Deploy vCluster with resource limits
2. Configure network policies
3. Set up ArgoCD for vCluster management

### Monitoring

1. Configure monitoring for DevPod usage
2. Set up alerts for resource limits
3. Monitor authentication success/failure rates

## Support

For issues with DevPod setup:
1. Check this documentation
2. Review DevPod logs: `devpod logs`
3. Contact your infrastructure team
4. Check the [DevPod documentation](https://devpod.sh/docs)