package sound

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"testing"
	"time"
)

// fakeOutput plays nothing: the test pulls a voice's frames as the system would.
type fakeOutput struct {
	voices    []*fakeVoice
	suspended bool
}

func (fake *fakeOutput) newVoice(source io.ReadSeeker) voice {
	made := &fakeVoice{source: source, volume: 1}
	fake.voices = append(fake.voices, made)
	return made
}
func (fake *fakeOutput) ready() <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}
func (fake *fakeOutput) err() error     { return nil }
func (fake *fakeOutput) suspend() error { fake.suspended = true; return nil }
func (fake *fakeOutput) resume() error  { fake.suspended = false; return nil }

type fakeVoice struct {
	source   io.ReadSeeker
	playing  bool
	volume   float64
	buffered int // bytes the "system" holds, not heard yet
	ended    bool
}

func (voice *fakeVoice) Play()                    { voice.playing = !voice.ended }
func (voice *fakeVoice) Pause()                   { voice.playing = false }
func (voice *fakeVoice) IsPlaying() bool          { return voice.playing }
func (voice *fakeVoice) SetVolume(volume float64) { voice.volume = volume }
func (voice *fakeVoice) BufferedSize() int        { return voice.buffered }
func (voice *fakeVoice) Close() error             { voice.playing = false; return nil }
func (voice *fakeVoice) Seek(offset int64, whence int) (int64, error) {
	voice.ended, voice.buffered = false, 0
	return voice.source.Seek(offset, whence)
}

// pull reads frames as the system would, and returns the stereo samples it got.
func (voice *fakeVoice) pull(t *testing.T, frames int) []float32 {
	t.Helper()
	bytes := make([]byte, frames*bytesPerFrame)
	read := 0
	for read < len(bytes) {
		count, err := voice.source.Read(bytes[read:])
		read += count
		if err == io.EOF {
			voice.ended, voice.playing = true, false
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	samples := make([]float32, read/4)
	for index := range samples {
		samples[index] = math.Float32frombits(binary.LittleEndian.Uint32(bytes[4*index:]))
	}
	return samples
}

// steps is a stream counting up one frame at a time: 0, 1, 2… in every channel.
type steps struct {
	format   Format
	length   int
	position int
}

func (counted *steps) Format() Format { return counted.format }
func (counted *steps) Read(samples []float32) (int, error) {
	written := 0
	for written+counted.format.Channels <= len(samples) && counted.position < counted.length {
		for channel := range counted.format.Channels {
			samples[written+channel] = float32(counted.position) + float32(channel)/10
		}
		written += counted.format.Channels
		counted.position++
	}
	if written == 0 {
		return 0, io.EOF
	}
	return written, nil
}
func (counted *steps) Rewind() error { counted.position = 0; return nil }

func TestAMonoSoundIsHeardOnBothSides(t *testing.T) {
	output := &fakeOutput{}
	player, err := newContextWith(output, 100).NewPlayer(&steps{format: Format{SampleRate: 100, Channels: 1}, length: 3})
	if err != nil {
		t.Fatal(err)
	}
	player.Play()
	samples := output.voices[0].pull(t, 10)
	want := []float32{0, 0, 1, 1, 2, 2}
	if len(samples) != len(want) {
		t.Fatalf("three frames, then the end: %v", samples)
	}
	for index := range want {
		if samples[index] != want[index] {
			t.Fatalf("left and right alike: %v", samples)
		}
	}
	if player.IsPlaying() {
		t.Fatal("it stops at its end")
	}
}

func TestASurroundSoundKeepsItsFrontPair(t *testing.T) {
	output := &fakeOutput{}
	player, _ := newContextWith(output, 100).NewPlayer(&steps{format: Format{SampleRate: 100, Channels: 6}, length: 2})
	player.Play()
	if samples := output.voices[0].pull(t, 2); samples[0] != 0 || samples[1] != 0.1 || samples[2] != 1 || samples[3] != 1.1 {
		t.Fatalf("the first two channels, left and right: %v", samples)
	}
}

func TestAnotherRateIsResampledSmoothly(t *testing.T) {
	output := &fakeOutput{}
	player, _ := newContextWith(output, 200).NewPlayer(&steps{format: Format{SampleRate: 100, Channels: 1}, length: 5})
	player.Play()
	samples := output.voices[0].pull(t, 100)
	left := make([]float32, 0, len(samples)/2)
	for index := 0; index < len(samples); index += 2 {
		left = append(left, samples[index])
	}
	want := []float32{0, 0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4}
	if len(left) != len(want) {
		t.Fatalf("twice the frames, as long as the sound lasts: %v", left)
	}
	for index := range want {
		if left[index] != want[index] {
			t.Fatalf("halfway frames between the source's: %v", left)
		}
	}
}

func TestALoopStartsAgainAndThePositionFollowsWhatIsHeard(t *testing.T) {
	output := &fakeOutput{}
	player, _ := newContextWith(output, 100).NewPlayer(&steps{format: Format{SampleRate: 100, Channels: 1}, length: 50})
	if err := player.SetLoop(true); err != nil {
		t.Fatal(err)
	}
	player.Play()
	voice := output.voices[0]
	voice.pull(t, 30)
	voice.buffered = 10 * bytesPerFrame // ten frames given, not heard yet
	if position := player.Position(); position != 200*time.Millisecond {
		t.Fatalf("thirty frames given, ten waiting: 20 frames heard, %v", position)
	}
	samples := voice.pull(t, 40) // past the end and round again
	if samples[2*19] != 49 || samples[2*20] != 0 || len(samples) != 80 {
		t.Fatalf("it goes on from its start: %v %v", samples[2*19], samples[2*20])
	}
	voice.buffered = 25 * bytesPerFrame // 20 frames of this time round and 5 of the last are waiting
	if position := player.Position(); position != 450*time.Millisecond {
		t.Fatalf("what is heard is near the end of the first time round: %v", position)
	}
}

func TestAPlayerRewindsPlaysAgainAndHasAVolume(t *testing.T) {
	output := &fakeOutput{}
	player, _ := newContextWith(output, 100).NewPlayer(&steps{format: Format{SampleRate: 100, Channels: 1}, length: 3})
	voice := output.voices[0]
	player.Play()
	voice.pull(t, 10)
	player.Play() // at its end: from its start
	if samples := voice.pull(t, 1); !voice.playing || samples[0] != 0 {
		t.Fatalf("an ended sound plays again from its start: %v", samples)
	}
	player.SetVolume(2)
	if voice.volume != 1 {
		t.Fatal("the volume goes up to 1")
	}
}

func TestStreamsThatCannotStartAgainDoNotLoop(t *testing.T) {
	output := &fakeOutput{}
	once := struct{ Stream }{Tone(440, time.Second, 1)} // hides Rewind
	player, _ := newContextWith(output, 48000).NewPlayer(once)
	if player.SetLoop(true) != ErrNotRewindable {
		t.Fatal("a stream that cannot rewind cannot loop")
	}
	if _, err := newContextWith(output, 48000).NewPlayer(&steps{}); err != ErrBadFormat {
		t.Fatal("a stream needs a format")
	}
}

func TestAToneRisesAndFallsWithoutAClick(t *testing.T) {
	note := Tone(1000, 100*time.Millisecond, 0.5)
	samples := make([]float32, 10000)
	count, _ := note.Read(samples)
	if count != 4800 {
		t.Fatalf("a tenth of a second at 48 kHz: %d", count)
	}
	loudest := float32(0)
	for _, sample := range samples[:count] {
		loudest = max(loudest, sample)
	}
	if samples[0] != 0 || abs(samples[count-1]) > 0.01 || loudest < 0.49 || loudest > 0.5 {
		t.Fatalf("silent at its ends, half as loud as it can be: %v %v %v", samples[0], samples[count-1], loudest)
	}
}

func abs(value float32) float32 { return max(value, -value) }

func TestSuspendAndResumeReachTheOutput(t *testing.T) {
	output := &fakeOutput{}
	context := newContextWith(output, 48000)
	context.Suspend()
	if !output.suspended {
		t.Fatal("suspended")
	}
	context.Resume()
	if output.suspended {
		t.Fatal("resumed")
	}
	select {
	case <-context.Ready():
	default:
		t.Fatal("ready")
	}
}

func TestAnOutputThatCannotStartIsKnownOnEverySystem(t *testing.T) {
	failure := fmt.Errorf("%w: no sound device", ErrOutputUnavailable)
	if !errors.Is(failure, ErrOutputUnavailable) {
		t.Fatal("an output that cannot start is not known by ErrOutputUnavailable")
	}
}
