// Package wav reads WAV files: the plainest sound files, samples one after another after a short
// header. Integer samples of 8, 16, 24 and 32 bits and floating point ones of 32 and 64 bits are
// read, in any number of channels.
package wav

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	sound "github.com/akrck02/go-sound"
)

// Stream is a WAV file's samples, read as they are played.
type Stream struct {
	source      io.Reader
	format      sound.Format
	floating    bool
	sampleBytes int
	dataStart   int64 // where the samples start, for a source that can seek
	dataLength  int64
	read        int64 // bytes of samples read
	bytes       []byte
}

// The WAVE format's codes for its sample encodings.
const (
	encodingInteger    = 1
	encodingFloat      = 3
	encodingExtensible = 0xfffe // the real code is the first two bytes of the sub-format
)

// ErrNotWAV is returned for a file that is not a WAV file.
var ErrNotWAV = errors.New("wav: not a WAV file")

// Decode reads a WAV file's header and returns its samples as a stream. A source that can seek (a
// file, a bytes.Reader of an embedded file) makes a stream that rewinds, so a player can loop it.
func Decode(source io.Reader) (*Stream, error) {
	var header [12]byte
	if _, err := io.ReadFull(source, header[:]); err != nil || string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return nil, ErrNotWAV
	}
	stream := &Stream{source: source}
	offset := int64(len(header))
	for {
		var chunk [8]byte
		if _, err := io.ReadFull(source, chunk[:]); err != nil {
			return nil, fmt.Errorf("wav: no samples: %w", err)
		}
		offset += int64(len(chunk))
		name, size := string(chunk[0:4]), int64(binary.LittleEndian.Uint32(chunk[4:]))
		if name == "data" {
			if stream.format.SampleRate == 0 {
				return nil, errors.New("wav: the samples come before their format")
			}
			stream.dataStart, stream.dataLength = offset, size
			return stream, nil
		}
		body := make([]byte, size+size%2) // chunks are padded to even sizes
		if _, err := io.ReadFull(source, body); err != nil {
			return nil, fmt.Errorf("wav: a chunk is cut short: %w", err)
		}
		offset += int64(len(body))
		if name == "fmt " {
			if err := stream.readFormat(body[:size]); err != nil {
				return nil, err
			}
		}
	}
}

// readFormat reads the format chunk: the encoding, the channels, the sample rate, the bits.
func (stream *Stream) readFormat(body []byte) error {
	if len(body) < 16 {
		return errors.New("wav: the format chunk is too short")
	}
	encoding := binary.LittleEndian.Uint16(body[0:])
	if encoding == encodingExtensible && len(body) >= 26 {
		encoding = binary.LittleEndian.Uint16(body[24:])
	}
	channels := int(binary.LittleEndian.Uint16(body[2:]))
	rate := int(binary.LittleEndian.Uint32(body[4:]))
	bits := int(binary.LittleEndian.Uint16(body[14:]))
	switch {
	case encoding == encodingInteger && (bits == 8 || bits == 16 || bits == 24 || bits == 32):
	case encoding == encodingFloat && (bits == 32 || bits == 64):
		stream.floating = true
	default:
		return fmt.Errorf("wav: samples encoded as %d in %d bits are not read", encoding, bits)
	}
	if channels == 0 || rate == 0 {
		return errors.New("wav: no channels or no sample rate")
	}
	stream.format, stream.sampleBytes = sound.Format{SampleRate: rate, Channels: channels}, bits/8
	return nil
}

// Format is the file's sample rate and channels.
func (stream *Stream) Format() sound.Format { return stream.format }

// Frames is how many frames the file holds.
func (stream *Stream) Frames() int64 {
	return stream.dataLength / int64(stream.sampleBytes*stream.format.Channels)
}

// Read fills samples with whole frames, from -1 to 1.
func (stream *Stream) Read(samples []float32) (int, error) {
	frameBytes := stream.sampleBytes * stream.format.Channels
	frames := min(int64(len(samples)/stream.format.Channels), (stream.dataLength-stream.read)/int64(frameBytes))
	if frames <= 0 {
		return 0, io.EOF
	}
	want := int(frames) * frameBytes
	if cap(stream.bytes) < want {
		stream.bytes = make([]byte, want)
	}
	bytes := stream.bytes[:want]
	got, err := io.ReadFull(stream.source, bytes)
	got -= got % frameBytes
	stream.read += int64(got)
	for index := 0; index < got/stream.sampleBytes; index++ {
		samples[index] = stream.sample(bytes[index*stream.sampleBytes:])
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		err = io.EOF // a file cut short ends where it is cut
	}
	return got / stream.sampleBytes, err
}

// sample reads one sample as a number from -1 to 1.
func (stream *Stream) sample(bytes []byte) float32 {
	switch {
	case stream.floating && stream.sampleBytes == 8:
		return float32(math.Float64frombits(binary.LittleEndian.Uint64(bytes)))
	case stream.floating:
		return math.Float32frombits(binary.LittleEndian.Uint32(bytes))
	}
	switch stream.sampleBytes {
	case 1: // eight-bit samples have no sign: 128 is silence
		return (float32(bytes[0]) - 128) / 128
	case 2:
		return float32(int16(binary.LittleEndian.Uint16(bytes))) / (1 << 15)
	case 3:
		value := int32(uint32(bytes[0])<<8|uint32(bytes[1])<<16|uint32(bytes[2])<<24) >> 8
		return float32(value) / (1 << 23)
	}
	return float32(int32(binary.LittleEndian.Uint32(bytes))) / (1 << 31)
}

// Rewind goes back to the first sample, when the source can seek.
func (stream *Stream) Rewind() error {
	seeker, can := stream.source.(io.Seeker)
	if !can {
		return sound.ErrNotRewindable
	}
	if _, err := seeker.Seek(stream.dataStart, io.SeekStart); err != nil {
		return err
	}
	stream.read = 0
	return nil
}
