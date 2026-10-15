package sound

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"sync"
	"time"
)

// Player plays one stream: paused when made, played, paused and played again from where it was,
// once or over and over, at a volume. A program keeps one for each sound it plays at the same
// time; the context mixes them.
type Player struct {
	voice  voice
	reader *voiceReader
	rate   int
}

// NewPlayer makes a player of a stream, paused at its start.
func (context *Context) NewPlayer(stream Stream) (*Player, error) {
	if !stream.Format().valid() {
		return nil, ErrBadFormat
	}
	reader := &voiceReader{conversion: newConverted(stream, context.sampleRate)}
	return &Player{voice: context.output.newVoice(reader), reader: reader, rate: context.sampleRate}, nil
}

// Play plays from where the player is; a sound that ended plays from its start.
func (player *Player) Play() {
	if player.reader.hasEnded() {
		player.Rewind()
	}
	player.voice.Play()
}

// Pause stops where the sound is.
func (player *Player) Pause() { player.voice.Pause() }

// IsPlaying tells whether the sound is playing: false once paused, or at its end.
func (player *Player) IsPlaying() bool { return player.voice.IsPlaying() }

// SetVolume sets how loud the sound plays, from 0 (silent) to 1 (as it was made).
func (player *Player) SetVolume(volume float64) { player.voice.SetVolume(min(max(volume, 0), 1)) }

// SetLoop makes the sound start again at its end, or stop there. A stream that cannot start again
// cannot loop.
func (player *Player) SetLoop(loop bool) error {
	if _, can := player.reader.conversion.source.(Rewinder); loop && !can {
		return ErrNotRewindable
	}
	player.reader.setLoop(loop)
	return nil
}

// Rewind goes back to the sound's start, dropping what was ready to be heard.
func (player *Player) Rewind() error {
	_, err := player.voice.Seek(0, io.SeekStart)
	return err
}

// Position is how far into the sound what is heard now is: what was read minus what still waits
// in the system's buffers. A film's pictures follow it.
func (player *Player) Position() time.Duration {
	waiting := int64(player.voice.BufferedSize() / bytesPerFrame)
	return durationOf(player.reader.heardFrames(waiting), player.rate)
}

// Close stops the sound and lets go of it. A closed player is not used again.
func (player *Player) Close() error { return player.voice.Close() }

// bytesPerFrame is a stereo frame of two float32 samples, as the output reads it.
const bytesPerFrame = 8

// voiceReader turns a converted stream into the bytes a voice reads, from the output's own
// goroutine, and counts the frames it gave, for the player's position.
type voiceReader struct {
	mutex      sync.Mutex
	conversion *converted
	samples    []float32
	loop       bool
	given      int64 // frames given since the sound last started
	lastLength int64 // frames of the last time round, when looping: what is heard may still be in it
	ended      bool
}

func (reader *voiceReader) Read(bytes []byte) (int, error) {
	reader.mutex.Lock()
	defer reader.mutex.Unlock()
	frames := len(bytes) / bytesPerFrame
	if frames == 0 {
		return 0, nil
	}
	if cap(reader.samples) < 2*frames {
		reader.samples = make([]float32, 2*frames)
	}
	samples := reader.samples[:2*frames]
	written := 0
	for written < len(samples) {
		count, err := reader.conversion.Read(samples[written:])
		written += count
		reader.given += int64(count / 2)
		if err == nil {
			continue
		}
		if !errors.Is(err, io.EOF) {
			return reader.encode(bytes, samples[:written]), err
		}
		if !reader.loop || reader.conversion.rewind() != nil {
			reader.ended = true
			return reader.encode(bytes, samples[:written]), io.EOF
		}
		reader.lastLength, reader.given = reader.given, 0
	}
	return reader.encode(bytes, samples), nil
}

// encode writes samples as little endian float32 bytes and says how many bytes it wrote.
func (reader *voiceReader) encode(bytes []byte, samples []float32) int {
	for index, sample := range samples {
		binary.LittleEndian.PutUint32(bytes[4*index:], math.Float32bits(sample))
	}
	return 4 * len(samples)
}

// Seek starts the sound again from its beginning: the only place a player seeks to.
func (reader *voiceReader) Seek(offset int64, whence int) (int64, error) {
	reader.mutex.Lock()
	defer reader.mutex.Unlock()
	if offset != 0 || whence != io.SeekStart {
		return 0, errors.New("sound: a player seeks to its start only")
	}
	if err := reader.conversion.rewind(); err != nil {
		return 0, err
	}
	reader.given, reader.lastLength, reader.ended = 0, 0, false
	return 0, nil
}

func (reader *voiceReader) setLoop(loop bool) {
	reader.mutex.Lock()
	defer reader.mutex.Unlock()
	reader.loop = loop
}

func (reader *voiceReader) hasEnded() bool {
	reader.mutex.Lock()
	defer reader.mutex.Unlock()
	return reader.ended
}

// heardFrames is how many frames into the sound is heard now, with some still waiting in the
// output's buffers. Just after a loop starts again, what is heard is the end of the last time round.
func (reader *voiceReader) heardFrames(waiting int64) int64 {
	reader.mutex.Lock()
	defer reader.mutex.Unlock()
	heard := reader.given - waiting
	if heard < 0 && reader.lastLength > 0 {
		heard += reader.lastLength
	}
	return max(heard, 0)
}
