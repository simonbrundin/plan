## Why
The current development setup requires developers to run Docker locally, which consumes significant laptop resources and creates inconsistent environments. DevPod with vCluster integration will offload all development workloads to the Kubernetes cluster while providing secure authentication through Authentik OIDC, eliminating the need for local kubeconfig files and ensuring proper isolation between developers.

## What Changes
- Integrate DevPod with existing vCluster infrastructure for isolated development environments
- Configure Authentik OIDC authentication to replace local kubeconfig distribution
- Create automated vCluster provisioning per developer
- Update development documentation and CI/CD pipelines
- Implement RBAC controls through Authentik groups

## Impact
- Affected specs: development-server
- Affected code: New DevPod configuration files and vCluster templates
- Breaking changes: None - this adds new capability alongside existing local development
- Security improvement: Removes need for local kubeconfig distribution
- Performance improvement: Offloads Docker workloads from developer laptops
