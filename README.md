`acmoi`
=======

My acme tools. General documentation on [godoc](https://godoc.org/github.com/whilp/acmoi). Commands:

-	[Define](https://godoc.org/github.com/whilp/acmoi/cmd/Define)
-	[Open](https://godoc.org/github.com/whilp/acmoi/cmd/Open)
-	[Grep](https://godoc.org/github.com/whilp/acmoi/cmd/Grep)
-	[Watch](https://godoc.org/github.com/whilp/acmoi/cmd/Watch)

To catch them all:

```bash
go get -u github.com/whilp/acmoi/...
```

The tools here do as little as possible to interact with acme before delegating work to external commands. Examples for these commands can be found in the `scripts/` directory:

```
scripts/acme-build
scripts/acme-check
scripts/acme-define
scripts/acme-format
scripts/acme-grep
scripts/acme-root
scripts/acme-test
```

These commands are typically invoked with the full path to a file as the first argument, so it is possible to switch behavior based on the file extension or other indicator. For example, `acme-format` runs [markdownfmt](https://github.com/shurcooL/markdownfmt) for files that end in `.md` and `go fmt` for files that end in `.go`.

TODO
----

-	make 2-1 chording work on OSX ([this](https://groups.google.com/forum/#!topic/comp.os.plan9/aEwQNcr80cQ) doesn't yet seem to do the trick)
-	Doc/acme-doc (with gogetdoc)
-	Test/acme-test
-	Build/acme-build
-	Manifest/acme-manifest (git ls-files)
-	Commit/acme-commit (git commit -v, w/ diff line numbers, perhaps separate windows for message and diff)
-	Add/acme-add (git add -f)
-	Diff (include plumber-compatible line numbers in git diff output)
-	scripts/acme-* support for other formats/languages
-	general: shorten files in Errors output relative to Root
