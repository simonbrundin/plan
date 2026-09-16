# Agent Preferences

This file contains preferences and requirements for AI coding agents working on this project.

## Runtime Preferences

### Nuxt Application
- **Use Bun for runtime**, NOT Node.js
- The Nuxt app uses `nitro preset: "node-server"` (for the server-side code), but runs on Bun (`oven/bun:1-slim` container)
- Dockerfile MUST use `oven/bun:1-slim` for production runtime, not `node:lts-alpine`

## Important Context

### Authentication Fix (fix-goals-401-auth-error)
- Zitadel returns opaque/JWE tokens that the JWT library can't parse
- Solution: Use session tokens with `introspectToken()` for Zitadel introspection
- Session tokens use HMAC-SHA256 signing with `SESSION_SECRET`
- Users auto-provisioned via `LookupOrCreateUser()` on first login

### Current Production Issues
- plan.simonbrundin.com not responding to HTTP requests (upstream timeout)
- Readiness probe failing: `/api/health` endpoint not responding inside Kubernetes
- Bun runtime may have issues with HTTP responses in production
- GitHub Actions cache causing stale builds with missing dependencies

## Secrets
- SESSION_SECRET: `3eDqjso1WefD4doxeG2nR4ZQvg4SjOMe669ac4B8UQE=`
- Zitadel CLIENT_SECRET: `pba!gvj5dqk7hut@NYR` (in 1Password item "Auth - Zitadel")

## Cluster Access
- Use certificate login: `KUBECONFIG=/tmp/kubeconfig-certificate kubectl...`
- Vault is down (no pods in vault namespace) - ESO can't sync secrets
- Auth secret patched manually as workaround
