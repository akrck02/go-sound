# 🔊 Using go-sound

## A context

A program opens the speakers once, with a context. The system allows one per program; a second
`NewContext` says `ErrContextExists`.

```go
context, err := sound.NewContext(sound.Options{Name: "Lighthouses"})
```

| Option | What it sets |
|---|---|
| `SampleRate` | How many frames a second are played: 48000 by default, the rate phones and most sound cards run at |
| `Latency` | How much sound is kept ready ahead of what is heard; 0 leaves it to the system. Shorter answers sooner, too short crackles |
| `Name` | The name a desktop's volume mixer shows |

`NewContext` returns at once. In a web page the browser lets sound start only after the person's
first touch or key: `context.Ready()` is closed once it can. `context.Err()` tells what went wrong
with the output (no sound device, a device unplugged).

## Players

A player plays one stream. Make one for each sound that may play at the same time: the context
mixes them.

```go
note, err := context.NewPlayer(sound.Tone(660, 600*time.Millisecond, 0.5))
note.Play()
```

| Call | What it does |
|---|---|
| `Play`, `Pause`, `IsPlaying` | Plays from where it is (from the start once it has ended), stops there, and tells which |
| `SetVolume(volume)` | From 0, silent, to 1, as the sound was made |
| `SetLoop(true)` | Starts again at its end, for music and the sea; a stream that cannot start again says `ErrNotRewindable` |
| `Rewind` | Back to its start, dropping what was ready to be heard |
| `Position` | How far into the sound what is heard now is: what was read, less what still waits in the system's buffers |
| `Close` | Stops and lets go of it |

A stream of any sample rate and any number of channels plays: it is turned into stereo at the
context's rate as it is read. A mono sound is heard on both sides, the first two channels of a
surround one are kept, and other rates are resampled between neighbouring frames.

## Streams

A stream gives frames of samples from -1 to 1. Decode a file with `wav.Decode` or `vorbis.Decode`
([sound files](formats.md)), make a beep with `sound.Tone`, or write one:

```go
type Stream interface {
	Format() sound.Format              // the sample rate and the channels
	Read(samples []float32) (int, error) // whole frames; io.EOF at the end
}
```

A stream that also has `Rewind() error` can loop. A decoder's stream rewinds when its source can
seek: a file, or a `bytes.Reader` of an embedded file.

## On a phone

A phone's applications go quiet when they leave the screen. Suspend the context then, and resume it
when the program comes back. With go-ui:

```go
options.Hidden = func() { context.Suspend() }
options.Shown = func() { context.Resume() }
```

On Android the sound goes through Oboe, which oto carries: AAudio from Android 11, OpenSL ES before
(AAudio did not always say when a phone switched from its speaker to headphones). Building for
Android compiles that C++ with the NDK, so cgo needs both `CC` and `CXX`; go-ui-apk sets both.

## On the Nintendo 3DS

Built with TinyGo for the `nintendo3ds` target, a context plays through the console's DSP. oto
does not reach the 3DS, so go-sound mixes its players itself (the same mixer the tests try) and
hands the DSP four buffers of a thirtieth of a second in turn.

- **The DSP needs its firmware**, `/3ds/dspfirm.cdc` on the SD card, which homebrew tools dump
  from the console. Without it the context opens all the same and `Err` says why it is silent:
  show it, so the player knows.
- **Yield once a frame.** TinyGo's scheduler there is cooperative, and the buffers are filled from
  a goroutine: go-ui's window yields every frame; a loop of its own calls `runtime.Gosched()`.
- **Choose the sounds' rate** as the context's (22050 frames a second is plenty for the console's
  speakers): nothing is resampled, and the processor has less to mix.
- Ogg Vorbis music decodes as it plays, at full speed on a New 3DS; WAV suits short effects.
