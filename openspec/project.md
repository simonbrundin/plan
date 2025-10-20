# Project Context

## Purpose
A goal management and planning application that helps users organize and track their objectives. The application supports hierarchical goal structures, user-specific goals, team collaboration, and progress tracking through a modern web interface.

## Tech Stack
- **Frontend**: Nuxt.js 3, Vue 3 (Composition API), TypeScript, Tailwind CSS
- **Backend**: Hasura GraphQL API, PostgreSQL database
- **ORM**: Drizzle ORM for database schema management
- **State Management**: Pinia stores
- **Testing**: Vitest (unit tests), Cucumber.js (BDD tests)
- **Deployment**: Kubernetes, Docker
- **Package Management**: Bun (primary), npm
- **CI/CD**: GitHub Actions

## Project Conventions

### Code Style
- **TypeScript**: Strict mode enabled, explicit types for all parameters and return values, avoid `any` type
- **Vue Components**: Use Composition API with `<script setup>`, PascalCase component names
- **Naming**: camelCase for variables/functions, kebab-case for files, PascalCase for Vue components
- **Imports**: ES6 modules - framework imports first, then external dependencies, then local imports
- **Error Handling**: Use try-catch blocks for async operations, throw descriptive Error objects
- **Comments**: DO NOT add comments unless explicitly requested
- **Formatting**: Follow existing code patterns, use Tailwind CSS utility classes

### Architecture Patterns
- **Full-Stack Separation**: Frontend (Nuxt.js) and backend (Hasura) in separate directories
- **GraphQL API**: Hasura provides GraphQL API layer over PostgreSQL
- **Database-First**: Schema defined in Drizzle ORM with migrations
- **Component Architecture**: Reusable Vue components with composition functions
- **State Management**: Pinia stores for client-side state
- **Middleware**: Nuxt middleware for authentication and routing

### Testing Strategy
- **BDD Tests**: Cucumber.js for behavior-driven development tests (`bun run test`)
- **Unit Tests**: Vitest for component and utility testing (`npx vitest run <test-file>`)
- **Test Coverage**: Focus on critical user paths and component logic
- **Test Organization**: Tests co-located with source files where appropriate

### Git Workflow
- **Branching**: Feature branches from main, semantic versioning
- **CI/CD**: GitHub Actions workflows for frontend and backend
- **Release**: Automated semantic releases using `.releaserc.json`
- **PR Process**: Pull request templates and automated checks

## Domain Context
Goal management application with hierarchical relationships between objectives. Users can create personal goals, establish parent-child relationships between goals, track completion status, and collaborate in teams. The system supports searching goals, managing user-specific goal assignments, and visualizing goal hierarchies.

## Important Constraints
- **Deployment**: Must run in Kubernetes environments with production overlays
- **Database**: PostgreSQL with schema migrations managed through Drizzle
- **Authentication**: Integrated with Authentik for user management
- **API**: GraphQL-first architecture with Hasura as the API layer

## External Dependencies
- **Hasura**: GraphQL engine and API layer
- **PostgreSQL**: Primary database
- **Authentik**: Authentication and user management system
- **Kubernetes**: Container orchestration for deployment
