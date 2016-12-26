# `acmoi`

My acme tools.

## `Do`

`Do` watches the acme log for `put` events. When it sees one, it invokes a series of scripts in the following order:

- `acme-format`
- `acme-check`
- `acme-build`
- `acme-test`

Each script is invoked in the root of the git repository containing whichever file acme just `put`, with the path to that file as its sole argument. `acme-format` is expected to rewrite the source file if needed; after it exits without error, its containing window will be issued a `get` command.