## Why
Developers need a consistent, reproducible development environment that integrates Kubernetes and Tilt for running the development server and services. A devcontainer that automatically sets up and runs the Tiltfile via Kubernetes will provide a seamless development experience without requiring manual setup of Kubernetes clusters and Tilt.

## What Changes
- Add a new devcontainer configuration that includes Kubernetes tools and Tilt
- Configure the devcontainer to automatically start Tilt with the Kubernetes-based Tiltfile.devpod
- Ensure all development services (PostgreSQL, Hasura, frontend) run properly within the containerized environment

## Impact
- Affected specs: development-server
- Affected code: New .devcontainer.json configuration, potential updates to Tiltfile.devpod
- No breaking changes to existing functionality</content>
</xai:function_call">Now let me create the tasks.md file. This should be an ordered list of small, verifiable work items. Based on the request, I need to:

1. Create a devcontainer configuration that includes Kubernetes tools
2. Configure it to run Tilt with the Kubernetes tiltfile
3. Test that the development environment works properly

Let me break this down into specific tasks. 

<xai:function_call name="write">
<parameter name="filePath">openspec/changes/create-devcontainer-kubernetes-tilt/tasks.md