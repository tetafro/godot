# Integration test

Main integration test for linter.

- If line contains `[DEFAULT]` string, it should be caught as error by the
linter with the setting `All: false`.
- If line contains `[ALL]` string, it should be caught as error by the
linter with the setting `All: true`.
- If line contains `[PASS]` string, it shouldn't be caught.
