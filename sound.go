// Package sound plays sound: effects, music and the sound of films, mixed and sent to the
// speakers on Android, Linux, macOS, Windows and in a web page.
//
// A program makes one Context, with NewContext, and a Player for each sound it plays. A Player
// reads a Stream: samples decoded from a file (packages wav and vorbis), made by the program
// (Tone), or a film's track. Streams of any sample rate and number of channels are played: the
// Context turns them into its own format as it reads them.
//
// Underneath, ebitengine/oto sends the mixed samples to the system: AAudio or OpenSL ES on
// Android (through Oboe), PulseAudio or ALSA on Linux, AudioToolbox on macOS, WASAPI on Windows and
// an AudioWorklet in a browser.
package sound

import (
	"errors"
	"fmt"
	"time"
)

// Format is how a stream's samples come: how many frames a second, and how many channels each
// frame holds, side by side (left then right for stereo).
type Format struct {
	SampleRate int
	Channels   int
}

// Stream is sound to play: frames of samples from -1 to 1, the channels of a frame side by side.
type Stream interface {
	Format() Format
	// Read fills samples with whole frames and says how many samples it wrote. At the end it
	// returns io.EOF, with the last samples or none.
	Read(samples []float32) (int, error)
}

// Rewinder is a stream that can start again from its beginning: a looping player rewinds it.
type Rewinder interface {
	Rewind() error
}

// ErrNotRewindable is a player's answer when asked to rewind or loop a stream that cannot start
// again (one read from a network).
var ErrNotRewindable = errors.New("sound: the stream cannot start again")

// ErrOutputUnavailable is the error when the system's sound output cannot start: no sound device,
// or a console's sound processor without its firmware. Its message says which; a program without
// sound still runs, and checks for it with errors.Is.
var ErrOutputUnavailable = errors.New("sound: the system's sound output cannot start")

// outputUnavailable is ErrOutputUnavailable with what made the output fail: the system's own
// error, or what is missing.
func outputUnavailable(detail string) error {
	return fmt.Errorf("%w: %s", ErrOutputUnavailable, detail)
}

// ErrBadFormat is returned for a stream whose format has no sample rate or no channels.
var ErrBadFormat = errors.New("sound: a stream needs a sample rate and at least one channel")

func (format Format) valid() bool { return format.SampleRate > 0 && format.Channels > 0 }

// framesIn is how many frames last a duration at a sample rate.
func framesIn(duration time.Duration, sampleRate int) int64 {
	return int64(duration) * int64(sampleRate) / int64(time.Second)
}

// durationOf is how long frames last at a sample rate.
func durationOf(frames int64, sampleRate int) time.Duration {
	return time.Duration(frames * int64(time.Second) / int64(sampleRate))
}
