# Integration test

Integration test for `getComments` method.

- If line contains `[DEFAULT]` string, it should be extracted by `getComments`
with `all == false` or `all == true` argument.
- If line contains `[ALL]` string, it should be extracted by `getComments`
only with `all == true` argument.
- If line contains `[NONE]` string, it should never be extracted by
`getComments`.
