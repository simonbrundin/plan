# Plan - Goal Management Application

A goal management and planning application that helps users organize and track
their objectives using a modern web interface with hierarchical goal structures.

## Tech Stack

- **Frontend**: Nuxt 4, TypeScript, Tailwind CSS
- **Backend**: PostgreSQL database
- **Testing**: Vitest, Cucumber.js
- **Deployment**: Kubernetes

## Agent Platform

Plan connects each goal to one or more agent runs. An `AgentRun` is the durable
record of that work and is stored in the `agent_runs` table with its phase,
status, summary, optional artifact, cost, and conversation ID. Keeping the run
in Plan means the interface can show the agent's progress even when the agent
service is temporarily unavailable.

### Run lifecycle

```mermaid
sequenceDiagram
    actor User as User
    participant Plan as Plan
    participant Worker as Agent worker
    participant Repo as Repository

    User->>Plan: Start an agent run for a goal
    Plan->>Worker: Queue an agent job
    Worker->>Worker: Run in an isolated workspace
    Worker->>Plan: Request approval for a consequential change
    User->>Plan: Approve or reject in the interface
    alt Approved
        Worker->>Repo: Apply changes and publish a branch
        Worker->>Plan: Move the run to verify
        User->>Plan: Review and merge the branch
        Plan->>Plan: Complete the run
    else Rejected
        Worker->>Plan: Block or fail the run
    end
```

The worker applies changes only in an isolated workspace. Changes that require
approval pause the job until a person decides in the Plan interface. An
approved change is applied in that workspace, tested, and published on a branch;
the branch is left for human review and merge. The agent job finishes directly
when the run has completed.

A run moves through the phases `define`, `plan`, `execute`, `verify`, and
`completed`. Its status is one of `queued`, `running`, `blocked`, `done`, or
`failed`. During verification, the agent presents evidence and artifacts for
review rather than merging the change itself.

## Authentication

The session lives in the **Go API** - not in Nuxt. Go creates, validates and
deletes the session and sets the cookies; Nuxt is a thin proxy in front of it.

```mermaid
sequenceDiagram
    autonumber
    actor B as Browser
    participant N as Nuxt (Nitro)
    participant G as Go API
    participant Z as Zitadel
    participant D as PostgreSQL

    B->>N: GET /auth/login
    N->>G: GET /api/v1/auth/login
    G-->>N: 307 to Zitadel + Set-Cookie oauth_state
    N-->>B: 307 to Zitadel (oauth_state)
    B->>Z: GET /oauth/v2/authorize (PKCE)
    Z-->>B: 302 to /api/v1/auth/callback?code=...
    B->>G: GET /api/v1/auth/callback
    G->>D: create user + session
    G-->>B: 307 to app root + Set-Cookie plan_session, plan_user
    B->>N: GET /api/proxy/goals (Cookie: plan_session)
    N->>G: GET /api/v1/goals (Cookie: plan_session)
    G->>D: validate session
    G-->>N: 200 OK
    N-->>B: 200 OK
```

```mermaid
flowchart LR
    B[Browser] -->|HTTP| N["Nuxt SPA<br/>(Nitro server)"]
    N -->|"proxy: Cookie + Set-Cookie"| G["Go API<br/>(Gin)"]
    G -->|SQL| D[("PostgreSQL<br/>(CNPG)")]
    G -->|OAuth 2.0 / PKCE| Z[Zitadel]
    G -->|session token| B
```

Rules:

- **Go owns the session.** `plan_session` (httpOnly) is set by the Go API in
  the login response - the session token never travels in a URL.
- **Nuxt is a thin proxy** (`server/utils/goApi.ts`): it forwards the `Cookie`
  header and copies back `Set-Cookie` and `Location`. It never follows
  redirects itself, otherwise the OAuth redirect would be resolved server-side
  instead of in the browser.
- **One session, two ways in**: the `plan_session` cookie for browsers,
  `Authorization: Bearer <token>` for API clients.
- **`plan_user`** is a readable hint cookie (sub + email) so the UI can render
  immediately. It is not authoritative - the server validates the session on
  every request.
- **Logout** (`POST /api/v1/auth/logout`) deletes the session server-side and
  clears the cookies.
- **Dev login** (`POST /api/v1/auth/dev-login`) creates a session for
  `dev-user` without Zitadel. It only exists when `DEV_AUTH_ENABLED=true`.

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