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
    users {
        int id PK
        string sub UK NN
        string email UK NN
        string first_name
        string last_name
        timestamp created NN
    }
    goals {
        int id PK
        string title
        string icon
        timestamp created NN
        timestamp started
        timestamp finished
        int inbox NN
    }
    user_goals {
        int user_id FK NN
        int goal_id FK NN
    }
    goal_relations {
        int parent_id FK NN
        int child_id FK NN
        int order NN
        int weight NN
    }
    goal_dependencies {
        int id PK
        int goal_id FK NN
        int depends_on_id FK NN
        timestamp created NN
    }
    users ||--o{ user_goals : ""
    goals ||--o{ user_goals : ""
    goals ||--o{ goal_relations : "parent"
    goals ||--o{ goal_relations : "child"
    goal_dependencies }o--|| goals : "waits for"
    goals }o--|| goal_dependencies : "blocks"
```
