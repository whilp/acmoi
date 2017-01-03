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

Most of the tools rely on the concept of a project. Each acme window contains either a file or a directory with a name; that name is a path, and that path is expected to exist within some project structure (typically a git repository). The `acme-root` command is invoked in the parent directory of the window's path and should return a path containing the window's name. Other commands are usually run from that root; within acme, output is usually routed to an `+Errors` file opened for a `.guide` file at the project's root. This collects all output from commands run on files within a project in a single window. The `.guide` file can be used to store notes; the output of `git ls-files` for easy access; 'bookmarks' using acme's search addresses (`README.md:/TODO`); or handy commands (`git add -e`, `acme-grep TODO`).

TODO
----

-	make 2-1 chording work on OSX ([this](https://groups.google.com/forum/#!topic/comp.os.plan9/aEwQNcr80cQ) doesn't yet seem to do the trick)
-	github stuff (hub)
-	Diff (include plumber-compatible line numbers in git diff output)
-	Doc/acme-doc (with gogetdoc, godef)
-	Test/acme-test
-	Build/acme-build
-	Manifest/acme-manifest (git ls-files)
-	Commit/acme-commit (git commit -v, w/ diff line numbers, perhaps separate windows for message and diff)
-	Add/acme-add (git add -f)
-	scripts/acme-* support for other formats/languages
