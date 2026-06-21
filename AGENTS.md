# Agent Guidelines for Plan Repository

## Commands

- **Build**: `cd frontend && bun run build`
- **Dev**: `cd frontend && bun run dev --host`
- **Test**: `cd frontend && bun run test` (Cucumber.js BDD tests)
- **Single test**: `cd frontend && npx vitest run <test-file>` (Vitest unit
  tests)
- **Database**: `cd frontend && bun run db:push` (push schema),
  `bun run db:migrate` (migrations)

## Code Style

- **TypeScript**: Strict mode, explicit types for parameters/returns, avoid
  `any`
- **Vue**: Composition API with `<script setup>`, PascalCase components
- **Naming**: camelCase variables, kebab-case files, PascalCase Vue components
- **Imports**: ES6 modules - framework imports first, then dependencies, then
  local
- **Error Handling**: try-catch for async operations, descriptive Error objects
- **Comments**: DO NOT ADD comments unless explicitly requested
- **Formatting**: Follow existing patterns, use Tailwind CSS classes
