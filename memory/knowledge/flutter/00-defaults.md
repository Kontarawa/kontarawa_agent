# Flutter defaults (starter)

## Goal

Keep UI simple, isolate state, make features easy to test.

## Project shape (starter)

- `features/<feature>/` with:
  - `ui/`
  - `state/` (controller/bloc/notifier)
  - `data/` (api, repository, models)

## Rules of thumb

- UI widgets should be dumb; state lives in one place.
- Prefer explicit navigation + typed routes (when possible).
- Avoid leaking platform channels into UI; wrap them in services.
