## ADDED Requirements
### Requirement: DevPod vCluster Integration
The development environment MUST support DevPod workspaces running in isolated vClusters within the main Kubernetes cluster, providing each developer with a dedicated virtual Kubernetes environment that mirrors production while maintaining security through Authentik OIDC authentication.

#### Scenario: DevPod workspace creation with OIDC authentication
- **GIVEN** a developer with Authentik credentials
- **WHEN** they run `devpod up` with a repository URL
- **THEN** DevPod authenticates via Authentik OIDC
- **AND** creates an isolated vCluster for the developer
- **AND** provisions a development container within the vCluster
- **AND** establishes secure port forwarding to the developer's machine

#### Scenario: Resource isolation and security
- **GIVEN** multiple developers using DevPod simultaneously
- **WHEN** each developer creates their workspace
- **THEN** each gets a separate vCluster with resource limits
- **AND** network policies prevent cross-developer communication
- **AND** RBAC controls limit actions within each vCluster
- **AND** audit logs track all DevPod activities

#### Scenario: Development workflow compatibility
- **GIVEN** a DevPod workspace in a vCluster
- **WHEN** a developer runs development commands
- **THEN** all existing workflows work (bun run dev, bun run test, etc.)
- **AND** hot reloading functions through port forwarding
- **AND** database connections work via Kubernetes services
- **AND** Tilt can manage local development services

#### Scenario: Cost optimization and resource management
- **GIVEN** DevPod workspaces running in vClusters
- **WHEN** developers are inactive
- **THEN** vClusters can be paused to save resources
- **AND** resource usage is monitored and reported
- **AND** automatic cleanup removes unused workspaces
- **AND** cost allocation tracks usage per developer/team
