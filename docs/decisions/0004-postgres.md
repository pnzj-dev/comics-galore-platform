# ADR 0004 – Database: PostgreSQL + Encore Migrations + SQLC

## Status
Accepted

## Context
We need a reliable relational store for users, tiers, comics, social graphs, quotas and payments.

## Decision
- PostgreSQL as the only database.
- Schema changes exclusively via **Encore SQL migrations**.
- Query layer: **SQLC** (preferred) for type-safe Go code.
- No ORM that generates raw SQL at runtime outside migrations.

## Consequences
- All tables, indexes and constraints are defined in migration files.
- Business logic never embeds raw SQL strings.
- Monthly quota calculations are simple, indexed aggregations on `download_logs`.
```
