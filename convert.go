package sound

import (
	"errors"
	"io"
)

// A context plays stereo frames at its own sample rate. A stream of another format is converted
// as it is read: its channels are mapped to two (a mono sound is heard on both sides, the first two
// channels of a surround one are kept), then its frames are resampled to the context's rate by
// linear interpolation between neighbouring frames, which keeps speech and music clean between the
// usual rates (44.1 and 48 kHz) at little cost.

// converted reads a stream as stereo frames at a sample rate.
type converted struct {
	source Stream
	format Format // the source's
	rate   int    // the frames' rate wanted
	step   float64
	place  float64       // where the next frame falls between previous and next, from 0 to 1
	frames [2][2]float32 // the source frames either side of it, in stereo
	filled int           // 2 once both are read, 0 at the start
	held   bool          // the source ended: its last frame is held for as long as it lasts
	// the source is read a block at a time: decoders pay for each call
	block    []float32
	consumed int // samples of block used
}

// blockFrames is how many frames are read from a source at once.
const blockFrames = 1024

func newConverted(source Stream, rate int) *converted {
	format := source.Format()
	return &converted{source: source, format: format, rate: rate, step: float64(format.SampleRate) / float64(rate)}
}

// Read fills samples with stereo frames at the wanted rate.
func (conversion *converted) Read(samples []float32) (int, error) {
	written := 0
	for written+2 <= len(samples) {
		if err := conversion.fill(); err != nil {
			if written > 0 && errors.Is(err, io.EOF) {
				return written, nil
			}
			return written, err
		}
		previous, next := conversion.frames[0], conversion.frames[1]
		weight := float32(conversion.place)
		samples[written] = previous[0] + (next[0]-previous[0])*weight
		samples[written+1] = previous[1] + (next[1]-previous[1])*weight
		written += 2
		conversion.place += conversion.step
	}
	return written, nil
}

// fill reads source frames until the next frame wanted lies between the two it keeps. At the
// source's end its last frame is held for as long as a frame lasts, so a sound plays as long at any
// rate; then it says io.EOF.
func (conversion *converted) fill() error {
	if conversion.filled == 0 {
		first, err := conversion.readFrame()
		if err != nil {
			return err
		}
		second, err := conversion.readFrame()
		if errors.Is(err, io.EOF) {
			second = first // a sound of one frame
		} else if err != nil {
			return err
		}
		conversion.frames, conversion.filled = [2][2]float32{first, second}, 2
	}
	for conversion.place >= 1 {
		next, err := conversion.readFrame()
		if errors.Is(err, io.EOF) && !conversion.held {
			conversion.held, next = true, conversion.frames[1]
		} else if err != nil {
			return err
		}
		conversion.place--
		conversion.frames[0], conversion.frames[1] = conversion.frames[1], next
	}
	return nil
}

// readFrame reads one source frame as stereo.
func (conversion *converted) readFrame() ([2]float32, error) {
	channels := conversion.format.Channels
	if conversion.consumed+channels > len(conversion.block) {
		if err := conversion.readBlock(); err != nil {
			return [2]float32{}, err
		}
	}
	frame := conversion.block[conversion.consumed : conversion.consumed+channels]
	conversion.consumed += channels
	if channels == 1 {
		return [2]float32{frame[0], frame[0]}, nil
	}
	return [2]float32{frame[0], frame[1]}, nil
}

// readBlock reads the source's next block of whole frames.
func (conversion *converted) readBlock() error {
	channels := conversion.format.Channels
	if cap(conversion.block) < blockFrames*channels {
		conversion.block = make([]float32, blockFrames*channels)
	}
	for attempt := 0; ; attempt++ {
		count, err := conversion.source.Read(conversion.block[:blockFrames*channels])
		count -= count % channels // a stream gives whole frames; a broken one loses its last part
		conversion.block, conversion.consumed = conversion.block[:count], 0
		if count > 0 {
			return nil
		}
		if err != nil {
			return err
		}
		if attempt == 100 {
			return io.ErrNoProgress
		}
	}
}

// rewind starts again from the source's beginning.
func (conversion *converted) rewind() error {
	rewinder, can := conversion.source.(Rewinder)
	if !can {
		return ErrNotRewindable
	}
	if err := rewinder.Rewind(); err != nil {
		return err
	}
	conversion.place, conversion.filled, conversion.held = 0, 0, false
	conversion.block, conversion.consumed = conversion.block[:0], 0
	return nil
}
