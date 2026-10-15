// Package vorbis reads Ogg Vorbis files: compressed sound, free of patents, small for music and
// effects alike. The decoding is jfreymuth/oggvorbis's, in Go.
package vorbis

import (
	"io"

	sound "github.com/akrck02/go-sound"
	"github.com/jfreymuth/oggvorbis"
)

// Stream is an Ogg Vorbis file's samples, decoded as they are played.
type Stream struct {
	reader   *oggvorbis.Reader
	seekable bool
}

// Decode reads an Ogg Vorbis file's headers and returns its samples as a stream. A source that can
// seek (a file, a bytes.Reader of an embedded file) makes a stream that rewinds, so a player can
// loop it.
func Decode(source io.Reader) (*Stream, error) {
	reader, err := oggvorbis.NewReader(source)
	if err != nil {
		return nil, err
	}
	_, seekable := source.(io.Seeker)
	return &Stream{reader: reader, seekable: seekable}, nil
}

// Format is the file's sample rate and channels.
func (stream *Stream) Format() sound.Format {
	return sound.Format{SampleRate: stream.reader.SampleRate(), Channels: stream.reader.Channels()}
}

// Frames is how many frames the file holds, when it says.
func (stream *Stream) Frames() int64 { return stream.reader.Length() }

// Read fills samples with whole frames, from -1 to 1.
func (stream *Stream) Read(samples []float32) (int, error) {
	channels := stream.reader.Channels()
	return stream.reader.Read(samples[:len(samples)-len(samples)%channels])
}

// Rewind goes back to the first sample, when the source can seek.
func (stream *Stream) Rewind() error {
	if !stream.seekable {
		return sound.ErrNotRewindable
	}
	return stream.reader.SetPosition(0)
}
