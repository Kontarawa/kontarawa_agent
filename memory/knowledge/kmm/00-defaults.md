# KMM defaults (starter)

## Goal

Ship features with predictable structure and minimal platform-specific code.

## Project shape (starter)

- Shared module exposes **use-cases** and **models**.
- Platform apps wire DI + UI.

## Rules of thumb

- Keep shared APIs simple and stable.
- Prefer DTO mapping at boundaries.
- If a library is painful on iOS: isolate behind an interface in shared.
