<div align="center">

# go-sound

**Sound for Go programs, on phones, desktops and the web.** Effects, music and, soon, the sound of
films: mixed and sent to the speakers through each system's own audio, with WAV and Ogg Vorbis read
in Go.

[![Go](https://img.shields.io/github/go-mod/go-version/akrck02/go-sound)](go.mod)
![Platforms](https://img.shields.io/badge/platforms-Android%20%7C%20Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20web-informational)
[![Licence](https://img.shields.io/github/license/akrck02/go-sound)](LICENSE)

</div>

## 🤖 AI disclosure

I write most of my code by hand, so I want to be clear about this project: **it was written by an
AI assistant (Claude, by Anthropic) under my direction, in October 2026.** I decided what to build
and how it should behave; the code, the tests and the docs were written by the AI. Every commit made
this way carries a `Co-Authored-By: Claude` line.

**Status: pending thorough review.** The code compiles and its tests pass, and on the Android 14
emulator the system reports its sound playing, pausing in the background and playing again, but I
have not yet reviewed it line by line. Treat it as a working draft, not as code I vouch for.

---

## 🚀 Quick start

```go
context, err := sound.NewContext(sound.Options{Name: "Lighthouses"})
…
//go:embed foghorn.ogg
var foghorn []byte

stream, err := vorbis.Decode(bytes.NewReader(foghorn))
player, err := context.NewPlayer(stream)
player.SetLoop(true)
player.Play()
```

On a phone, stop the sound when the program leaves the screen and play it again when it comes back
(go-ui's `window.Options.Hidden` and `Shown` call `context.Suspend` and `context.Resume`).
[More on using it →](docs/using.md)

## 🔍 How it works

<p align="center">
  <img alt="wav.Decode, vorbis.Decode or sound.Tone → a player per sound, converted to stereo at the context's rate → oto, mixed, to each system's audio" src="docs/images/how-it-works.svg" width="880">
</p>

Every stream is converted as it is read, to stereo at the context's rate, and played by
[oto](https://github.com/ebitengine/oto), which mixes the players and talks to each system.
[More on the design →](docs/design.md)

## 📚 Documentation

| | |
|---|---|
| [Using go-sound](docs/using.md) | A context, players, looping, volume, the position heard, phones going to the background. |
| [Sound files](docs/formats.md) | WAV and Ogg Vorbis, and which to choose. |
| [The design](docs/design.md) | Why a repository of its own, why oto, how a film's pictures will follow its sound. |

## Licence

MIT, see [LICENSE](LICENSE).
