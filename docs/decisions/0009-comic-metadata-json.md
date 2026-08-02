# ADR 0009 – Comic Creation via Archive + Metadata JSON

## Status
Accepted

## Context
Uploaders need a fast way to publish comics, especially when they already have structured metadata.

## Decision
- Two creation modes: rich manual form **or** single archive upload.
- When using the archive mode, the archive must contain a documented metadata file (`comic.json` or `metadata.json`).
- The backend validates the JSON against a published JSON Schema, extracts media, and creates the comic.
- Invalid or incomplete metadata produces clear, actionable errors.

## Consequences
- A public JSON Schema is part of the documentation.
- Background job handles validation + extraction.
- Both modes produce identical comic records.
