package vorbis

import (
	"bytes"
	"io"
	"math"
	"os"
	"testing"

	sound "github.com/akrck02/go-sound"
)

// note.ogg is a quarter of a second of a 440 Hz note, mono at 44.1 kHz.
func TestAnOggVorbisNoteIsDecoded(t *testing.T) {
	data, err := os.ReadFile("testdata/note.ogg")
	if err != nil {
		t.Fatal(err)
	}
	stream, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if stream.Format() != (sound.Format{SampleRate: 44100, Channels: 1}) {
		t.Fatalf("mono at 44.1 kHz: %+v", stream.Format())
	}
	samples := decodeAll(t, stream)
	if len(samples) < 11000 || len(samples) > 11100 {
		t.Fatalf("a quarter of a second: %d samples", len(samples))
	}
	crossings := 0
	for index := 1; index < len(samples); index++ {
		if (samples[index-1] < 0) != (samples[index] < 0) {
			crossings++
		}
	}
	if math.Abs(float64(crossings)-220) > 10 {
		t.Fatalf("a 440 Hz note crosses zero 880 times a second: %d in a quarter", crossings)
	}
	if err := stream.Rewind(); err != nil {
		t.Fatal(err)
	}
	if again := decodeAll(t, stream); len(again) != len(samples) {
		t.Fatalf("rewound, it decodes again whole: %d", len(again))
	}
}

func decodeAll(t *testing.T, stream *Stream) []float32 {
	t.Helper()
	var all []float32
	buffer := make([]float32, 1000)
	for {
		count, err := stream.Read(buffer)
		all = append(all, buffer[:count]...)
		if err == io.EOF {
			return all
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestWhatIsNotVorbisSaysSo(t *testing.T) {
	if _, err := Decode(bytes.NewReader([]byte("RIFF....WAVE"))); err == nil {
		t.Fatal("not an Ogg Vorbis file")
	}
}
