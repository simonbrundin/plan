## Why
The existing "simon dev" command attempts to use DevPod with the Kubernetes provider, but the configuration is incomplete. The .devpod.yaml file is missing proper Kubernetes provider configuration, and the devcontainer setup doesn't integrate with the existing Tilt infrastructure. This prevents developers from using the command to start their development environment in Kubernetes.

## What Changes
- Fix the existing "simon dev" command to properly work with DevPod Kubernetes provider
- Create proper .devpod.yaml configuration for Kubernetes integration
- Ensure devcontainer integrates with existing Tilt setup for development services
- Update devcontainer configuration to include necessary tools and startup scripts

## Impact
- Affected specs: New devpod-integration capability
- Affected code: .devpod.yaml, .devcontainer.json, and simon CLI workflow script
- No breaking changes to existing functionality