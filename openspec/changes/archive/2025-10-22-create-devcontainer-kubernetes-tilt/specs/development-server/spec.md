## ADDED Requirements
### Requirement: DevContainer Kubernetes Tilt Integration
The repository MUST provide a devcontainer configuration that automatically sets up a development environment with Kubernetes and Tilt integration. The devcontainer MUST include all necessary tools (kubectl, helm, tilt) and MUST automatically start the Tiltfile.devpod when the container starts, providing a complete development environment with PostgreSQL, Hasura, and the frontend application running in Kubernetes.

#### Scenario: DevContainer startup with Kubernetes and Tilt
- **WHEN** a developer opens the repository in a devcontainer
- **THEN** the container includes kubectl, helm, and tilt binaries
- **AND** the container has access to a Kubernetes cluster
- **AND** Tilt automatically starts with Tiltfile.devpod
- **AND** PostgreSQL and Hasura services are deployed to Kubernetes
- **AND** the frontend development server is accessible

#### Scenario: Development workflow in devcontainer
- **WHEN** working within the devcontainer
- **THEN** all development commands work (bun run dev, bun run test, bun run db:push)
- **AND** hot reloading functions correctly
- **AND** database connections are properly configured
- **AND** port forwarding works for accessing services

#### Scenario: Tilt resource management
- **WHEN** Tilt is running in the devcontainer
- **THEN** all Tilt resources start successfully (postgres, hasura, bun dev)
- **AND** port forwards are established (5432, 8080, 3000)
- **AND** the development server is accessible at localhost:3000</content>
</xai:function_call">Now let me validate the proposal to make sure it follows the OpenSpec format correctly. 

<xai:function_call name="bash">
<parameter name="command">cd /Users/simon/repos/plan && openspec validate create-devcontainer-kubernetes-tilt --strict