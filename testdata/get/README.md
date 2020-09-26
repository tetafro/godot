# Integration test

Integration test for `getComments` method.

- If line contains `[DECL]` string, it should be extracted by `getComments`
with `scope` argument one of: `DeclScope`, `TopLevelScope`, `AllScope`.
- If line contains `[TOP]` string, it should be extracted by `getComments`
with `scope` argument one of: `TopLevelScope`, `AllScope`.
- If line contains `[ALL]` string, it should be extracted by `getComments`
with `scope` argument `AllScope`.
