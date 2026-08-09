# Plan - Goal Management Application

A goal management and planning application that helps users organize and track
their objectives using a modern web interface with hierarchical goal structures.

## Tech Stack

- **Frontend**: Nuxt 4, TypeScript, Tailwind CSS
- **Backend**: PostgreSQL database
- **Testing**: Vitest, Cucumber.js
- **Deployment**: Kubernetes

## Database Schema

```mermaid
erDiagram
    USERS {
        INTEGER id PK
        VARCHAR(255) sub
        VARCHAR(40) email
        VARCHAR(16) first_name
        VARCHAR(40) last_name
        TIMESTAMPTZ created
    }
    GOALS {
        INTEGER id PK
        VARCHAR(64) title
        VARCHAR(100) icon
        TIMESTAMPTZ created
        TIMESTAMPTZ started
        TIMESTAMPTZ finished
        INTEGER inbox
    }
    USERGOALS {
        INTEGER user_id FK
        INTEGER goal_id FK
    }
    GOALRELATIONS {
        INTEGER parent_id FK
        INTEGER child_id FK
        INTEGER order
        INTEGER weight
    }
    GOALDEPENDENCIES {
        INTEGER goal_id FK
        INTEGER depends_on_id FK
        TIMESTAMPTZ created
    }
    USERS ||--o{ USERGOALS : "has"
    GOALS ||--o{ USERGOALS : "assigned to"
    GOALS ||--o{ GOALRELATIONS : "parent"
    GOALS ||--o{ GOALRELATIONS : "child"
    GOALS ||--o{ GOALDEPENDENCIES : "depends on"
```
