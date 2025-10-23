## ADDED Requirements
### Requirement: DevPod Kubernetes Provider Integration
The existing "simon dev" command MUST be fixed to properly configure and use the Kubernetes provider for DevPod, enabling developers to start their devcontainer in a Kubernetes environment with all necessary development services.

#### Scenario: DevPod command execution with Kubernetes provider
- **WHEN** a developer runs the "simon dev" command
- **THEN** DevPod uses the Kubernetes provider with proper configuration
- **AND** the devcontainer starts in the configured Kubernetes namespace
- **AND** Tilt automatically deploys PostgreSQL, Hasura, and other development services

#### Scenario: Kubernetes environment and Tilt integration
- **WHEN** the DevPod workspace initializes
- **THEN** the .devpod.yaml configuration ensures proper Kubernetes context
- **AND** Tilt starts with Tiltfile.devpod for service management
- **AND** port forwarding is established (5432, 8080, 3000)
- **AND** the development environment is fully operational

#### Scenario: Development workflow in Kubernetes
- **WHEN** working within the DevPod Kubernetes environment
- **THEN** all development commands execute properly (bun run dev, bun run test, bun run db:push)
- **AND** hot reloading works for frontend development
- **AND** database connections use the Kubernetes services
- **AND** the application is accessible at localhost:3000