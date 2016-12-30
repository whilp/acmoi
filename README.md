# `acmoi`

My acme tools.

## `Watch`

`Do` watches the acme log for `put` events. When it sees one, it invokes a series of scripts in the following order:

- `acme-format`
- `acme-check`
- `acme-build`
- `acme-test`

Each script is invoked in the root of the git repository containing whichever file acme just `put`, with the path to that file as its sole argument. `acme-format` is expected to rewrite the source file if needed; after it exits without error, its containing window will be issued a `get` command.

When invoked without the `-daemon` option, `Do` will run once in the context of the current acme window and then exit.

## TODO

- Watch
- Grep/acme-grep
- Doc/acme-doc (with gogetdoc)
- Test/acme-test
- Build/acme-build
- Manifest/acme-manifest (git ls-files)
- general: shorten files in Errors output relative to Root