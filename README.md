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
        int id PK
        string sub
        string email
        string first_name
        string last_name
        timestamp created
    }
    GOALS {
        int id PK
        string title
        string icon
        timestamp created
        timestamp started
        timestamp finished
        int inbox
    }
    USERGOALS {
        int user_id FK
        int goal_id FK
        int PRIMARY_KEY user_id, goal_id
    }
    GOALRELATIONS {
        int parent_id FK
        int child_id FK
        int order
        int weight
        int PRIMARY_KEY parent_id, child_id
    }
    GOALDEPENDENCIES {
        int goal_id FK
        int depends_on_id FK
        timestamp created
        int PRIMARY_KEY goal_id, depends_on_id
    }
    USERS ||--o{ USERGOALS : ""
    GOALS ||--o{ USERGOALS : ""
    GOALS ||--o{ GOALRELATIONS : "parent"
    GOALS ||--o{ GOALRELATIONS : "child"
    GOALDEPENDENCIES }o--|| GOALS : "waits for"
    GOALS }o--|| GOALDEPENDENCIES : "blocks"
```
