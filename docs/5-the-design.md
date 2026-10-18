# 🧭 The design

## A repository of its own

Sound is its own package, depending on neither Go Renderer nor Go UI:

- **A program without sound pays nothing.** On Android, the output is C++ (Oboe) built with cgo; a
  program that plays nothing should not carry it.
- **It is useful beyond a window.** A tool with no window, or a game drawn with another engine, can
  play sound with it.
- **It has its own life.** An audio thread, suspending in the background, the phone's audio focus.

## The output: oto

[ebitengine/oto](https://github.com/ebitengine/oto) sends the sound to each system: it is what the
Ebitengine game engine plays through, and it is kept up to date.

| System | Through |
|---|---|
| Android | Oboe, which oto carries and links statically (nothing more goes in the package): AAudio from Android 11, OpenSL ES before, because AAudio did not always say when the sound moved from the speaker to headphones |
| Linux | PulseAudio, spoken in Go; ALSA when no PulseAudio runs, loaded when the program starts. No cgo |
| macOS | AudioToolbox |
| Windows | WASAPI |
| A web page | An AudioWorklet, started at the first touch or key, as browsers ask |
| Nintendo 3DS | The DSP, through libctru's ndsp, built with TinyGo: Go Sound mixes the players itself and hands the DSP buffers of 16-bit stereo frames |

Google advises Oboe or AAudio over OpenSL ES for new programs; Oboe makes that choice on each phone.
Writing AAudio by hand (300 lines of C, as Go Renderer's films are) would cover Android alone and
lose Oboe's fixes; miniaudio's Go binding (malgo) needs cgo everywhere and has no web output.

Go Sound keeps oto behind two small interfaces (`output` and `voice`): the tests put a fake output
there that pulls samples as the system would, and another output could take oto's place.

## How a sound plays

1. A **stream** gives frames of samples, in its own rate and channels.
2. The **player** converts them as they are read: channels to stereo, then frames to the context's
   rate, interpolating between neighbouring frames. A sound plays as long at any rate: its last frame
   is held for as long as it lasts.
3. **oto** reads the player's frames from its own goroutine, mixes the players at their volumes and
   hands them to the system.
4. The player counts the frames it gave. **What is heard now** is that count less what still waits in
   oto's and the system's buffers: `Player.Position`. Just after a loop starts again, what is heard
   is still the end of the time before.

## Films

A film's pictures play today and its sound does not: Go Renderer's Android films decode their
picture track only, and a browser's film is muted (browsers play no sound before a touch). The plan:

- **Go Renderer** decodes a film's sound track too, with a second of the phone's decoders
  (`MediaCodec`), and hands its samples out, playing nothing itself.
- **Go UI**'s `window.Video` gives those samples as a stream, and takes a clock: a small interface of
  its own, so Go UI need not know Go Sound.
- **The sound leads.** A picture can be dropped or shown twice and nobody notices; a gap in the sound
  is always heard. So a player's `Position` is the clock, and the film shows the picture due by then.
- **In a browser** the film's own element plays its sound, in step with its pictures; it is unmuted
  after the first touch.
- A small package joins the two where both are used (Go UI for Android, or a package here).
