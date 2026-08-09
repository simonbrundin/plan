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
        string firstName
        string lastName
        timestamp createdAt
    }
    GOALS {
        int id PK
        string title
        string icon
        timestamp createdAt
        timestamp startedAt
        timestamp finishedAt
        int inbox
    }
    USERGOALS {
        int userId FK
        int goalId FK
    }
    GOALRELATIONS {
        int parentId FK
        int childId FK
        int orderIndex
        int weight
    }
    GOALDEPENDENCIES {
        int id PK
        int goalId FK
        int dependsOnId FK
        timestamp createdAt
    }
    USERS ||--o{ USERGOALS : ""
    GOALS ||--o{ USERGOALS : ""
    GOALS ||--o{ GOALRELATIONS : "parent"
    GOALS ||--o{ GOALRELATIONS : "child"
    GOALDEPENDENCIES }o--|| GOALS : "waits for"
    GOALS }o--|| GOALDEPENDENCIES : "blocks"
```
