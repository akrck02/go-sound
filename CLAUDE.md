# go-sound

Sound for Go programs: a context that mixes players and sends them to the speakers on Android,
Linux, macOS, Windows and in a web page (through ebitengine/oto), streams of any format converted
as they are read, and decoders for WAV (`wav`) and Ogg Vorbis (`vorbis`). Start with `docs/README.md`.

## Commands

```bash
go vet ./...
go test ./...                      # no sound device needed: a fake output pulls the samples
# for Android, cgo builds oto's C++ (Oboe) with the NDK: CC and CXX both, as go-ui-apk sets them
```

## Rules

- **Generic only.** This project knows nothing about any particular world, game, company or story.
  Never put names, places, dates, numbers or settings from a user's own project in code, comments,
  tests, examples, docs or commit messages. Examples use made-up places (lighthouses on an island).
- **Readable code.** Full-word names (no acronyms, no one-letter variables), small functions that do
  one thing, early returns instead of nested branches, comments that say *why*. Match the code
  around you.
- **Tests with every change.** `go vet ./...` and `go test ./...` must pass before committing.
- **Docs live in `docs/`**, written for people: explanatory, friendly, with images where they help.
  Keep the README short and point to `docs/`. Update the docs in the same commit as the behaviour.
- **AI disclosure.** Commits written with an AI assistant end with a `Co-Authored-By: Claude …` line,
  and the *AI disclosure* section of the README stays true. The code is pending human review.
- **Git.** Commit only when asked; never push, force-push, rewrite history or delete branches without
  the author's explicit permission. Do not change the global git config.
- **Standalone.** This package depends on neither go-renderer nor go-ui; they do not depend on it.
  What joins them (a film's sound) is small and lives where both are used.

## Human and AI code

- Hand-written code goes to `main`. AI work goes on feature branches (`feature/…`) and is merged by
  the author after review.
