# Integration test

Main integration test for linter.

- If line contains `[DECL]` string, it should be caught as error by the
linter with the setting `Scope: DeclScope`.
- If line contains `[TOP]` string, it should be caught as error by the
linter with the setting `Scope: TopLevelScope`.
- If line contains `[ALL]` string, it should be caught as error by the
linter with the setting `Scope: AllScope`.
- If line contains `[PASS]` string, it shouldn't be caught.
