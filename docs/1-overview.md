<p align="center"><img alt="" src="images/logo.svg" width="120"></p>

# Welcome to go-sound 👋

**go-sound plays sound from Go programs**: beeps, music, the sea, a foghorn. On phones, computers, web pages and the Nintendo 3DS.

<p align="center"><img alt="The waveform of a short tone: it grows over its first milliseconds and fades over its last" src="images/hero-light.png" width="720"></p>

A program opens the speakers once, and plays as many sounds at once as it likes: notes it makes itself, or WAV and Ogg files it carries. go-sound mixes them and sends them to each system's own audio. On a phone it goes quiet when the app leaves the screen.

## Is it for me?

- ✅ You write Go and want **sound** in a program, a game or a go-ui app.
- ✅ You want to understand **what a sound is** for a computer: samples, rates, mixing.

## Where to start

Pick the line that sounds like you.

| If you… | Read | Time |
|---|---|---|
| are **new** and want to hear something | 🏴‍☠️ [The quest](2-the-quest.md): the sounds of a harbour, in 6 levels | 40 min |
| want to know **how to use it** in your program | [Using go-sound](3-using-go-sound.md) | 10 min |
| have a **sound file** and wonder which format | [Sound files](4-sound-files.md) | 5 min |
| want to know **why it is built this way** | [The design](5-the-design.md) | — |
| want to know **what is coming** | [Roadmap](6-roadmap.md) | — |

## All the pages, in reading order

The menu on the left (or the files in `docs/`) follows this order.

**Start**

- [Overview](1-overview.md): this page
- [The quest](2-the-quest.md): six levels, from a beep to an app

**Use**

- [Using go-sound](3-using-go-sound.md)
- [Sound files](4-sound-files.md)

**The project**

- [The design](5-the-design.md)
- [Roadmap](6-roadmap.md)

## 🚧 Not done yet

- The sound of films: go-renderer decodes a film's pictures only (see [the roadmap](6-roadmap.md)).
- MP3 and Opus files.
- Android's audio focus: lowering the sound when another application plays.
- Heard on a phone and seen playing on the Android 14 emulator; a desktop and a browser are built
  for but not heard yet.

## Words you will meet

| Word | What it means |
|---|---|
| Go | A programming language. To try these projects you install it once, from go.dev, and type `go run …` in a terminal. |
| Terminal | The window where you type commands: *Terminal* on macOS and Linux, *PowerShell* on Windows. |
| Repository | A project's folder of files, kept on GitHub. `git clone <address>` copies it to your computer. |
| Quest | The tutorial of these docs: short levels, each ending with something that works and an achievement. |

## The family

go-sound is one of a family of small projects that work together:

| Project | What it does |
|---|---|
| [go-renderer](https://github.com/akrck02/go-renderer) | draws 3D scenes and flat interfaces |
| [go-ui](https://github.com/akrck02/go-ui) | buttons, panels and whole applications |
| [go-ui-android](https://github.com/akrck02/go-ui-android) | Android apps that look like Android's own |
| **go-sound** (this one) | sound and music |
| [go-scene-builder](https://github.com/akrck02/go-scene-builder) | 3D scenes from a flat map |
| [go-simulator](https://github.com/akrck02/go-simulator) | simulations that change a scene |
| [go-simulator-terrain](https://github.com/akrck02/go-simulator-terrain) | ground shaped by rivers |
| [go-simulator-climate](https://github.com/akrck02/go-simulator-climate) | weather and biomes |
| [go-simulator-flora](https://github.com/akrck02/go-simulator-flora) | forests that grow |
| [go-simulator-urban](https://github.com/akrck02/go-simulator-urban) | towns that grow |
| [go-simulator-motion](https://github.com/akrck02/go-simulator-motion) | a scene that moves |
| [go-3ds](https://github.com/akrck02/go-3ds) | Go on the Nintendo 3DS |

> ⚠️ All of them are **educational experiments**, written with an AI assistant and pending review by their author: good for learning and playing, not for production.
