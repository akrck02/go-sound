# 🗺️ Roadmap

What is left, in the order it is best done. Phase 1 (this repository: a context, players, WAV and
Ogg Vorbis, suspend and resume) is done, and was heard on a phone in October 2026.

## Phase 2: the sound of films

A film's pictures play today (go-ui's `window.Video`) and its sound does not. See
[the design](5-the-design.md#films) for why it is split this way.

1. **go-renderer** (`android/video.c`): open the film's sound track with a second `AMediaCodec`
   (`audio/mp4a-latm` for AAC, `audio/opus`), from the same `AMediaExtractor` or a second one on the
   same file, and decode it a step at a time, as the pictures are. Hand out its samples (16-bit or
   float PCM, with the track's rate and channels from the output format) without playing them.
   Rewind and the film's end must reach the sound track too.
2. **go-ui** (`window/video.go`): give the samples as a stream (`Read(samples []float32)`, `Format`),
   in a small interface of go-ui's own, so go-ui does not import go-sound. Let `Video` take a clock
   (`func() time.Duration`) and show the picture due by then, dropping or repeating pictures; with
   no clock it keeps its own, as today.
3. **Joining them**: a small package (here as `sound/film`, or in go-ui-android) makes a go-sound
   player of the film's stream and gives the film the player's `Position` as its clock. Play, pause
   and rewind go to both.
4. **In a browser**: the film's `<video>` element plays its own sound in step with its pictures;
   unmute it after the person's first touch (it is muted now so it can start without one).
5. **Try it**: the gallery's film (`cmd/gallery/coast.mp4` has no sound track yet: make one with
   ffmpeg), on the emulator and a phone, checking the pictures keep up with the sound.

## Phase 3: more

- **Audio focus on Android**: ask for it when a sound starts (`AudioManager.requestAudioFocus`
  through JNI, in go-renderer), lower or pause the sound when another application takes it.
- **Volume by kind**: music and effects turned up and down apart.
- **MP3 and Opus**: go-mp3 is archived; look at goxiph, kazzmir/opus-go and darui3018823/opus, and
  check them against reference files before choosing.
- **A better resampler**: linear interpolation is fine between 44.1 and 48 kHz; far apart rates
  (8 kHz speech) would sound cleaner through a short windowed filter.
- **Hear the desktop and the web**: both are built but not heard yet (the containers here have no
  sound device). A go-ui example with a button, run on a desktop and in a browser.
