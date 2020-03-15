# godot

Linter that checks if all top-level comments contain a period at the
end of the last sentence if needed.

[CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments#comment-sentences) quote:

> Comments should begin with the name of the thing being described
> and end in a period

## Install and run

```sh
go get -u github.com/tetafro/godot/cmd/godot
godot ./myproject/main.go
```
