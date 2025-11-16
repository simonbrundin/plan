# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Structure

This is a full-stack goal management application with hierarchical goal tracking:

- **Frontend**: Nuxt 4 (Vue 3) app in `frontend/` directory with TypeScript, Tailwind CSS, and Nuxt UI
- **Backend**: Hasura GraphQL backend with Authentik OAuth integration for authentication
- **Database**: PostgreSQL with Drizzle ORM schema management
- **Environments**: Kubernetes configurations with Hasura, Tilt for local dev, ArgoCD for GitOps
- **Database Schema**: Available at [drawdb.app](https://www.drawdb.app/editor?shareId=a3185456f5496cf6a48840b0ecfea7e0)

## Key Commands

### Development

```bash
# Frontend development (from frontend/ directory)
cd frontend
bun run dev        # Start Nuxt dev server (use bun, not npm)
bun run build      # Build for production
bun run generate   # Generate static site
bun run preview    # Preview production build

# Testing
bun run test       # Run all tests with Cucumber.js
bun run test:watch # Watch mode for tests

# Database
bun run db:generate  # Generate Drizzle migrations
bun run db:migrate   # Run migrations
bun run db:push      # Push schema to database
bun run db:studio    # Open Drizzle Studio
```

### Testing Architecture

- **Cucumber.js** for BDD tests (primary test framework)
- Test files in `frontend/test/` and `frontend/features/`
- Watch mode available via `bun run test:watch`

## Frontend Architecture (Nuxt 4)

### Key Technologies

- **Nuxt 4** with Vue 3 Composition API
- **TypeScript** (strict mode)
- **Tailwind CSS 4** + **Nuxt UI** components
- **Shadcn/UI** components with custom component directory
- **Nuxt Auth Utils** for authentication
- **Reka UI** for additional Vue components
- **Bun** runtime (Nitro preset)

### Directory Structure

```
frontend/
├── app/
│   ├── app.vue              # Root application component
│   ├── pages/               # File-based routing
│   │   ├── index.vue        # Landing page
│   │   ├── root-goals.vue   # Top-level goals (no parents)
│   │   └── goal/[id].vue    # Individual goal detail page
│   ├── components/          # Vue components
│   ├── graphql/             # GraphQL query files (.gql)
│   ├── stores/              # Pinia stores (goals.ts for state management)
│   ├── plugins/             # Nuxt plugins (graphql-auth.ts for Hasura auth)
│   └── lib/                 # Utility libraries
├── server/
│   ├── api/auth/            # Server API routes for authentication
│   ├── routes/auth/         # OAuth routes (Authentik integration)
│   └── db/                  # Database utilities and Drizzle setup
├── drizzle/                 # Database migrations
└── nuxt.config.ts           # Nuxt configuration
```

### Code Conventions (from CRUSH.md)

- Use **Composition API** with `<script setup>` syntax
- **TypeScript**: Explicit types for function parameters and return values, avoid `any`
- **Naming**: PascalCase for Vue components, camelCase for variables, kebab-case for files
- **Imports**: ES6 modules with framework imports first, then dependencies, then local imports
- **Error Handling**: Try-catch blocks for async operations, descriptive Error objects

### UI Components

- **Shadcn/UI** components in `components/ui/` (no prefix)
- **Nuxt UI** components available globally
- **Lucide Vue** icons available

## GraphQL & Data Architecture

### Hasura Integration

- GraphQL endpoint: `http://localhost:8080/v1/graphql` (local dev)
- Production: `https://plan-hasura.simonbrundin.com/v1/graphql`
- Authentication handled via `app/plugins/graphql-auth.ts`:
  - Development: Uses admin secret (`dev-admin-secret`)
  - Production: JWT tokens from Authentik with Hasura claims
- GraphQL queries stored in `app/graphql/*.gql` files
- Code generation via `nuxt-graphql-client` (disabled during build)

### State Management

- **Pinia store** (`app/stores/goals.ts`) for client-side state
- Key entities:
  - `goals`: All goal objects with id, title, created, finished
  - `relations`: Parent-child relationships between goals (goal_relations table)
- Computed getters for root goals, parents, children, search
- Actions: loadGoals, addGoal, updateGoal, removeGoal, addRelation, removeRelation

### Database Schema

- **Goals**: Hierarchical goal system with parent-child relationships
- **User Goals**: Many-to-many relationship (user_goals junction table)
- **Authentication**: Users table with `sub` column for IdP integration (Authentik)
- Schema managed with Drizzle ORM in `frontend/drizzle/`

## Environment Configuration

Environment variables in `frontend/.env`:

- `GQL_HOST`: Hasura GraphQL endpoint
- `HASURA_GRAPHQL_ADMIN_SECRET`: Admin secret for development
- `BETTER_AUTH_URL`: Authentication base URL
- OAuth configuration for Authentik in `runtimeConfig`

## Deployment

- **Docker-based** with multi-platform builds (linux/arm64)
- **Kubernetes**: GitOps with ArgoCD, manifests in `environments/kubernetes/`
- **Local development**: Tilt workflow (if configured in `environments/local/`)
- **CI/CD**: GitHub Actions with semantic-release for versioning
- **Hasura**: Deployed as separate service, handles GraphQL backend
- Bun används istället för npm i frontend/

