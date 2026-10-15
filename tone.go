package sound

import (
	"io"
	"math"
	"time"
)

// Tone is a pure note: a sine wave at a frequency for a while, at a volume from 0 to 1. Its start
// and end fade over a few milliseconds, so it does not click. Tests use it, and a program can use
// it for a beep.
func Tone(frequency float64, duration time.Duration, volume float64) Stream {
	return &tone{frequency: frequency, volume: volume, length: framesIn(duration, toneRate)}
}

// toneRate is a tone's sample rate.
const toneRate = 48000

// toneFade is how long a tone takes to rise and to fall.
const toneFade = 5 * time.Millisecond

type tone struct {
	frequency, volume float64
	length, position  int64
}

func (note *tone) Format() Format { return Format{SampleRate: toneRate, Channels: 1} }

func (note *tone) Read(samples []float32) (int, error) {
	if note.position >= note.length {
		return 0, io.EOF
	}
	written := 0
	fade := float64(framesIn(toneFade, toneRate))
	for written < len(samples) && note.position < note.length {
		envelope := min(1, float64(note.position)/fade, float64(note.length-note.position)/fade)
		phase := 2 * math.Pi * note.frequency * float64(note.position) / toneRate
		samples[written] = float32(note.volume * envelope * math.Sin(phase))
		written++
		note.position++
	}
	return written, nil
}

func (note *tone) Rewind() error {
	note.position = 0
	return nil
}
