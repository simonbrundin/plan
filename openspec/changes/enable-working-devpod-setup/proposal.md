## Why
The project has DevPod infrastructure in place from previous changes, but DevPod is not currently working for developers. The existing setup has configuration gaps and missing components that prevent successful workspace creation and development workflow execution.

## What Changes
- Fix DevPod configuration to ensure proper workspace initialization
- Update devcontainer configuration for complete development environment
- Correct Tiltfile.devpod to work with current project structure
- Add missing Kubernetes resources and network policies
- Ensure proper database and Hasura connectivity in DevPod environments
- Update documentation with working setup instructions

## Impact
- Affected specs: development-server
- Affected code: .devpod.yaml, .devcontainer.json, Tiltfile.devpod, Kubernetes manifests
- Breaking changes: None - fixes existing non-functional setup
- Developer experience improvement: Enables working DevPod development environments