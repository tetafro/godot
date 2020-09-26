# Integration test

Integration test for `getComments` method.

- If line contains `[DECL]` string, it should be extracted by `getComments`
with `scope == DeclScope` or `scope == TopLevelScope` argument.
- If line contains `[TOP]` string, it should be extracted by `getComments`
only with `scope == TopLevelScope` argument.
- If line contains `[NONE]` string, it should never be extracted by
`getComments`.
