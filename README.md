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
        varchar sub UK NN
        varchar email UK NN
        varchar firstName
        varchar lastName
        timestamp createdAt NN
    }
    GOALS {
        int id PK
        varchar title
        varchar icon
        timestamp createdAt NN
        timestamp startedAt
        timestamp finishedAt
        int inbox NN
    }
    USERGOALS {
        int userId FK NN
        int goalId FK NN
    }
    GOALRELATIONS {
        int parentId FK NN
        int childId FK NN
        int orderIndex NN
        int weight NN
    }
    GOALDEPENDENCIES {
        int id PK
        int goalId FK NN
        int dependsOnId FK NN
        timestamp createdAt NN
    }
    USERS ||--o{ USERGOALS : ""
    GOALS ||--o{ USERGOALS : ""
    GOALS ||--o{ GOALRELATIONS : "parent"
    GOALS ||--o{ GOALRELATIONS : "child"
    GOALDEPENDENCIES }o--|| GOALS : "waits for"
    GOALS }o--|| GOALDEPENDENCIES : "blocks"
```
