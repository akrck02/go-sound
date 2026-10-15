package wav

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"testing"

	sound "github.com/akrck02/go-sound"
)

// file makes a WAV file: a format chunk, a chunk to skip, and the samples.
func file(encoding uint16, channels, rate, bits int, samples []byte) []byte {
	var format bytes.Buffer
	binary.Write(&format, binary.LittleEndian, encoding)
	binary.Write(&format, binary.LittleEndian, uint16(channels))
	binary.Write(&format, binary.LittleEndian, uint32(rate))
	binary.Write(&format, binary.LittleEndian, uint32(rate*channels*bits/8))
	binary.Write(&format, binary.LittleEndian, uint16(channels*bits/8))
	binary.Write(&format, binary.LittleEndian, uint16(bits))
	var body bytes.Buffer
	body.WriteString("WAVE")
	chunk := func(name string, data []byte) {
		body.WriteString(name)
		binary.Write(&body, binary.LittleEndian, uint32(len(data)))
		body.Write(data)
		if len(data)%2 == 1 {
			body.WriteByte(0)
		}
	}
	chunk("fmt ", format.Bytes())
	chunk("LIST", []byte("ISFT lighthouse")) // odd: padded
	chunk("data", samples)
	var whole bytes.Buffer
	whole.WriteString("RIFF")
	binary.Write(&whole, binary.LittleEndian, uint32(body.Len()))
	whole.Write(body.Bytes())
	return whole.Bytes()
}

func readAll(t *testing.T, stream *Stream) []float32 {
	t.Helper()
	var all []float32
	buffer := make([]float32, 3)
	for {
		count, err := stream.Read(buffer[:len(buffer)-len(buffer)%stream.Format().Channels])
		all = append(all, buffer[:count]...)
		if err == io.EOF {
			return all
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestEveryDepthReadsFromMinusOneToOne(t *testing.T) {
	half16 := make([]byte, 4)
	binary.LittleEndian.PutUint16(half16, uint16(int16(1<<14)))
	binary.LittleEndian.PutUint16(half16[2:], uint16(0x8000)) // -1
	float := make([]byte, 8)
	binary.LittleEndian.PutUint32(float, math.Float32bits(0.25))
	binary.LittleEndian.PutUint32(float[4:], math.Float32bits(-0.75))
	for name, test := range map[string]struct {
		data []byte
		want []float32
	}{
		"8 bits":   {file(encodingInteger, 1, 8000, 8, []byte{128, 192, 0}), []float32{0, 0.5, -1}},
		"16 bits":  {file(encodingInteger, 2, 44100, 16, half16), []float32{0.5, -1}},
		"24 bits":  {file(encodingInteger, 1, 48000, 24, []byte{0, 0, 0x40, 0, 0, 0xc0}), []float32{0.5, -0.5}},
		"float 32": {file(encodingFloat, 2, 48000, 32, float), []float32{0.25, -0.75}},
	} {
		stream, err := Decode(bytes.NewReader(test.data))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		got := readAll(t, stream)
		if len(got) != len(test.want) {
			t.Fatalf("%s: %v", name, got)
		}
		for index := range got {
			if math.Abs(float64(got[index]-test.want[index])) > 1e-6 {
				t.Fatalf("%s: %v, want %v", name, got, test.want)
			}
		}
	}
}

func TestAWAVRewindsWhenItsSourceSeeks(t *testing.T) {
	data := file(encodingInteger, 1, 8000, 8, []byte{128, 192, 64})
	stream, _ := Decode(bytes.NewReader(data))
	if stream.Format() != (sound.Format{SampleRate: 8000, Channels: 1}) || stream.Frames() != 3 {
		t.Fatalf("its format and length: %+v %d", stream.Format(), stream.Frames())
	}
	readAll(t, stream)
	if err := stream.Rewind(); err != nil {
		t.Fatal(err)
	}
	if again := readAll(t, stream); len(again) != 3 || again[1] != 0.5 {
		t.Fatalf("from its start again: %v", again)
	}
	once, _ := Decode(struct{ io.Reader }{bytes.NewReader(data)})
	if once.Rewind() != sound.ErrNotRewindable {
		t.Fatal("a source that cannot seek cannot rewind")
	}
}

func TestWhatIsNotAWAVSaysSo(t *testing.T) {
	if _, err := Decode(bytes.NewReader([]byte("OggS and more"))); err != ErrNotWAV {
		t.Fatal(err)
	}
	if _, err := Decode(bytes.NewReader(file(2, 1, 8000, 4, nil))); err == nil {
		t.Fatal("an encoding not read is said")
	}
}
