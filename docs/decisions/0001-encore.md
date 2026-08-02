# ADR 0001 – Backend Framework: Encore (Go)

## Status
Accepted

## Context
We need a backend that is:
- Type-safe
- Easy to deploy
- Provides built-in auth, background jobs, tracing and migrations
- Works well with a modern TypeScript frontend

## Decision
Use **Encore** with the Go runtime as the sole backend framework.

## Consequences
- Official `encore app create` must be used for scaffolding.
- All APIs are defined as Encore services.
- Database migrations are Encore SQL migrations.
- Background work uses Encore’s task system.
- The generated TypeScript client becomes the preferred way for the frontend to talk to the backend.
- We deliberately reject raw net/http, Gin, Echo, Fiber, etc.
```
