# ADR 0015 – Content Language & UI Internationalization

## Status
Accepted

## Context
Comics audiences are global (manga, manhwa, BD, webtoons). Each work has a natural language, and the product chrome should be localizable.

## Decision
1. Every comic has required `content_language` (ISO 639-1 / BCP 47 as needed).
2. UI is internationalized; default locale is **English (`en`)**.
3. Priority UI locales for engagement: en, ja, es, ko, fr, pt-BR, zh-CN, de, it, id.
4. UI locale (user preference) is independent of comic content language.
5. Admin can enable/disable locales and set defaults.
6. V1 ships English UI + content_language on comics; other locale packs can land progressively.

## Consequences
- List/search support language filters.
- Upload and metadata JSON must carry language.
- Translation workflow is required for each enabled locale pack.
