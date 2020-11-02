# Integration test

Main integration test for linter.

Tags:
- `[PERIOD_DECL]` - line should be caught as an error by the linter with the
settings `Scope: DeclScope, Period: true`.
- `[PERIOD_TOP]` - line should be caught as an error by the linter with the
settings `Scope: TopLevelScope, Period: true`.
- `[PERIOD_ALL]` - line should be caught as an error by the linter with the
settings `Scope: AllScope, Period: true`.
- `[CAPITAL_DECL]` - line should be caught as an error by the linter with the
settings `Scope: DeclScope, Capital: true`.
- `[CAPITAL_TOP]` - line should be caught as an error by the linter with the
settings `Scope: TopLevelScope, Capital: true`.
- `[CAPITAL_ALL]` - line should be caught as an error by the linter with the
settings `Scope: AllScope, Capital: true`.
- `[PASS]` - line shouldn't be caught.
