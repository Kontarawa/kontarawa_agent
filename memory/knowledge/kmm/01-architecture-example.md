# KMM architecture example (intentionally simple)

## Layers

- `domain/` — business rules (pure Kotlin)
- `data/` — networking + persistence + mapping
- `presentation/` — state + events (no UI toolkit types)

## Dumb example: feature "Login"

- **Domain**
  - `LoginUseCase(email, password) -> Result<User>`
- **Data**
  - `AuthApi.login(...) -> LoginDto`
  - `LoginDto -> User` mapper
- **Presentation**
  - `LoginState(email, password, isLoading, error)`
  - `LoginEvent.Submit`

## Notes

- Keep `presentation` independent from Compose/SwiftUI.
- UI layer only observes state and sends events.
