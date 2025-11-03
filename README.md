# Plan - Goal Management Application

A goal management and planning application that helps users organize and track
their objectives using a modern web interface with hierarchical goal structures.

## Tech Stack

- **Frontend**: Nuxt.js 3, Vue 3, TypeScript, Tailwind CSS
- **Backend**: Hasura GraphQL API, PostgreSQL database
- **ORM**: Drizzle ORM
- **Testing**: Vitest, Cucumber.js
- **Deployment**: Kubernetes, Docker

## Database Schema

[Database Schema](https://www.drawdb.app/editor?shareId=a3185456f5496cf6a48840b0ecfea7e0)

## Development Setup

### Option 1: DevPod with Kubernetes and Tilt (Recommended)

This project supports DevPod for reproducible development environments using
Kubernetes, vCluster, and Authentik OIDC authentication.

#### Prerequisites

1. Install [DevPod](https://devpod.sh/docs/getting-started/install)
2. Set up a Kubernetes cluster (local or remote)
3. Install [kubectl](https://kubernetes.io/docs/tasks/tools/)
4. Configure kubectl to access your cluster

#### Quick Start with DevPod

1. **Add Kubernetes Provider**:

   ```bash
   devpod provider add kubernetes
   ```

2. **Configure Provider** (if needed):

   ```bash
   devpod provider set-options kubernetes --option KUBERNETES_NAMESPACE=default
   ```

3. **Create Workspace**:

   ```bash
   devpod up .
   ```

4. **Start Development Environment**: DevPod will automatically:
   - Install dependencies
   - Set up the database
   - Start Tilt for local development services

#### Accessing Services

- **Frontend**: http://localhost:3000
- **Hasura Console**: http://localhost:8080/console
- **PostgreSQL**: localhost:5432

### Option 3: Secure DevPod with vCluster (Recommended)

For secure, isolated development environments using vCluster and Authentik OIDC:

#### Prerequisites

1. **DevPod CLI**: Install
   [DevPod](https://devpod.sh/docs/getting-started/install)
2. **Authentik Access**: Organization Authentik instance with DevPod OIDC
   provider configured
3. **vCluster Access**: Access to organization vCluster infrastructure

#### Setup

1. **Configure DevPod Provider**:

   ```bash
   devpod provider add kubernetes
   ```

2. **Set OIDC Environment Variables**:

   ```bash
   export KUBERNETES_OIDC_ISSUER_URL="https://authentik.example.com/application/o/devpod/"
   export KUBERNETES_OIDC_CLIENT_ID="your-client-id"
   export KUBERNETES_OIDC_CLIENT_SECRET="your-client-secret"
   ```

3. **Create Workspace**:
   ```bash
   devpod up .
   ```

#### Features

- **Isolated Environments**: Each developer gets their own vCluster namespace
- **OIDC Authentication**: Secure authentication through Authentik
- **Resource Limits**: Automatic resource quotas and network policies
- **Kubernetes Native**: Full Kubernetes API access within vCluster
- **Tilt Integration**: Local development services via Tilt

#### Accessing Services

- **Frontend**: http://localhost:3000
- **Hasura Console**: http://localhost:8080/console
- **PostgreSQL**: localhost:5432 (via port-forward)

### Option 2: Local Development with Tilt

For traditional local development:

1. **Prerequisites**:
   - Node.js LTS
   - Bun (recommended)
   - Docker and Docker Compose
   - Tilt

2. **Setup**:

   ```bash
   cd environments/local
   tilt up
   ```

3. **Access Services**:
   - Frontend: http://localhost:3000
   - Hasura Console: http://localhost:8080/console

### Option 3: Manual Setup

```bash
# Install dependencies
cd frontend
bun install

# Start database
cd environments/local
docker-compose up -d

# Push database schema
cd frontend
bun run db:push

# Start development server
bun run dev
```

## Development Environment Details

### DevPod Configuration

The `.devpod.yaml` file configures the development workspace:

- **Provider**: Kubernetes (runs containers in your K8s cluster)
- **IDE**: VS Code with essential extensions pre-installed
- **Dev Container**: Ubuntu-based with Node.js, Docker, and development tools
- **Services**: PostgreSQL, Hasura, and frontend development server

### Tilt Integration

Tilt manages the local development services defined in
`environments/local/Tiltfile`:

- **PostgreSQL**: Database with initialized schema and seed data
- **Hasura**: GraphQL API with console access
- **Frontend**: Nuxt.js development server with hot reload
- **Tests**: Frontend test runner in watch mode
- **Drizzle Studio**: Database management interface

### Available Commands

```bash
# Frontend commands (from frontend/ directory)
bun run dev          # Start development server
bun run build        # Build for production
bun run test         # Run tests
bun run db:push      # Push database schema
bun run db:studio    # Open Drizzle Studio

# Backend commands (via Tilt)
tilt up              # Start all services
tilt down            # Stop all services
```

### Environment Variables

- `NODE_ENV`: Set to "development"
- `PORT`: Frontend port (default: 3000)
- Database connection configured via Hasura environment variables

### Ports

- **3000**: Frontend development server
- **8080**: Hasura GraphQL API and console
- **5432**: PostgreSQL database

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js
  features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out
[the Next.js GitHub repository](https://github.com/vercel/next.js) - your
feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the
[Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme)
from the creators of Next.js.

Check out our
[Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying)
for more details.
