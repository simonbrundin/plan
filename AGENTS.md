<!-- OPENSPEC:START -->

# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:

- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:

- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# Agent Guidelines for Plan Repository

## Commands

- **Build**: `cd frontend && bun run build`
- **Dev**: `cd frontend && bun run dev --host`
- **Test**: `cd frontend && bun run test` (Cucumber.js BDD tests)
- **Single test**: `cd frontend && npx vitest run <test-file>` (Vitest unit tests)
- **Database**: `cd frontend && bun run db:push` (push schema), `bun run db:migrate` (migrations)

## Code Style

- **TypeScript**: Strict mode, explicit types for parameters/returns, avoid `any`
- **Vue**: Composition API with `<script setup>`, PascalCase components
- **Naming**: camelCase variables, kebab-case files, PascalCase Vue components
- **Imports**: ES6 modules - framework imports first, then dependencies, then local
- **Error Handling**: try-catch for async operations, descriptive Error objects
- **Comments**: DO NOT ADD comments unless explicitly requested
- **Formatting**: Follow existing patterns, use Tailwind CSS classes

