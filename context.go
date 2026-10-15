package sound

import (
	"errors"
	"io"
	"sync/atomic"
	"time"
)

// Context is a program's way to the speakers: it mixes its players at one sample rate, in
// stereo. The system allows one per program.
type Context struct {
	output     output
	sampleRate int
}

// Options are how a context plays.
type Options struct {
	// SampleRate is how many frames a second the context plays: 48000 by default, the rate phones
	// and most sound cards run at, so nothing is resampled twice.
	SampleRate int
	// Latency is how much sound is kept ready for the system ahead of what is heard: 0 leaves it to
	// the system. Shorter answers sooner (a game's effects); too short crackles.
	Latency time.Duration
	// Name is the program's name, which a desktop's volume mixer shows.
	Name string
}

// DefaultSampleRate is a context's sample rate unless Options say another.
const DefaultSampleRate = 48000

// ErrContextExists is NewContext's answer when the program already has a context.
var ErrContextExists = errors.New("sound: a program has one context; keep the first")

var contextMade atomic.Bool

// NewContext opens the system's sound output. It returns at once: on a web page the browser lets
// sound play only after the person's first touch or key, and Ready tells when it can.
func NewContext(options Options) (*Context, error) {
	if !contextMade.CompareAndSwap(false, true) {
		return nil, ErrContextExists
	}
	rate := options.SampleRate
	if rate <= 0 {
		rate = DefaultSampleRate
	}
	opened, err := openSystemOutput(rate, options.Latency, options.Name)
	if err != nil {
		contextMade.Store(false)
		return nil, err
	}
	return &Context{output: opened, sampleRate: rate}, nil
}

// newContextWith is a context on an output of the caller's: the tests' fake one.
func newContextWith(opened output, rate int) *Context {
	return &Context{output: opened, sampleRate: rate}
}

// SampleRate is how many frames a second the context plays.
func (context *Context) SampleRate() int { return context.sampleRate }

// Ready is closed once sound can play: at once on most systems, after the first touch or key on a
// web page.
func (context *Context) Ready() <-chan struct{} { return context.output.ready() }

// Err is what went wrong with the output, if anything did (no sound device, the device lost).
func (context *Context) Err() error { return context.output.err() }

// Suspend stops all sound, as a program going to the background should: on a phone, call it when
// the program leaves the screen.
func (context *Context) Suspend() error { return context.output.suspend() }

// Resume plays again what Suspend stopped.
func (context *Context) Resume() error { return context.output.resume() }

// output is where a context's sound goes: the system's (oto), or the tests' fake one.
type output interface {
	newVoice(source io.ReadSeeker) voice
	ready() <-chan struct{}
	err() error
	suspend() error
	resume() error
}

// voice is one sound the output plays, reading its frames as stereo float32 samples in little
// endian bytes: an oto.Player.
type voice interface {
	Play()
	Pause()
	IsPlaying() bool
	SetVolume(volume float64)
	BufferedSize() int
	Seek(offset int64, whence int) (int64, error)
	Close() error
}
