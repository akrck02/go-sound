package sound

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"time"
)

// steady is a source of stereo float32 frames that all hold one value, for so many frames.
func steady(value float32, frames int) *bytes.Reader {
	data := make([]byte, frames*bytesPerFrame)
	for index := 0; index < frames*2; index++ {
		binary.LittleEndian.PutUint32(data[4*index:], math.Float32bits(value))
	}
	return bytes.NewReader(data)
}

func TestTheMixerAddsVoicesAtTheirVolumes(t *testing.T) {
	var mixing mixer
	first, second := mixing.newVoice(steady(0.25, 64)), mixing.newVoice(steady(0.5, 64))
	first.Play()
	second.Play()
	second.SetVolume(0.5)
	frames := make([]int16, 2*16)
	mixing.mixInto(frames)
	half := float32(0.5) // 0.25 + 0.5 × 0.5
	want := int16(half * math.MaxInt16)
	if frames[0] != want || frames[31] != want {
		t.Fatalf("mixed %d and %d, want %d", frames[0], frames[31], want)
	}
}

func TestTheMixerClipsInsteadOfWrappingAround(t *testing.T) {
	var mixing mixer
	for range 3 {
		mixing.newVoice(steady(0.8, 8)).Play()
	}
	frames := make([]int16, 2*4)
	mixing.mixInto(frames)
	if frames[0] != math.MaxInt16 {
		t.Fatalf("three loud voices mixed to %d, want the loudest sample", frames[0])
	}
}

func TestAVoiceStopsAtItsEndAndSilenceFollows(t *testing.T) {
	var mixing mixer
	voice := mixing.newVoice(steady(0.5, 4))
	voice.Play()
	frames := make([]int16, 2*8)
	mixing.mixInto(frames)
	if voice.IsPlaying() || frames[2*3] == 0 || frames[2*4] != 0 {
		t.Fatalf("playing %v; frames %v: want four frames of sound, then silence", voice.IsPlaying(), frames)
	}
}

func TestWhatWaitsInTheBuffersIsCountedUntilPlayed(t *testing.T) {
	var mixing mixer
	voice := mixing.newVoice(steady(0.5, 64))
	voice.Play()
	mixing.mixInto(make([]int16, 2*16))
	mixing.mixInto(make([]int16, 2*16))
	if voice.BufferedSize() != 32*bytesPerFrame {
		t.Fatalf("%d bytes waiting, want two buffers' worth", voice.BufferedSize())
	}
	mixing.played(16)
	if voice.BufferedSize() != 16*bytesPerFrame {
		t.Fatalf("%d bytes waiting after one buffer played, want one", voice.BufferedSize())
	}
}

func TestAPlayerOnAMixerPlaysItsStream(t *testing.T) {
	var mixing mixer
	context := newContextWith(&mixerOutput{&mixing}, 8000)
	player, err := context.NewPlayer(Tone(440, time.Second, 0.5))
	if err != nil {
		t.Fatal(err)
	}
	player.Play()
	frames := make([]int16, 2*64)
	mixing.mixInto(frames)
	loud := false
	for _, sample := range frames {
		loud = loud || sample > 1000 || sample < -1000
	}
	if !loud {
		t.Fatal("a tone played through the mixer is silent")
	}
}

// mixerOutput is a mixer as an output, for a player on it.
type mixerOutput struct{ *mixer }

func (output *mixerOutput) ready() <-chan struct{} {
	opened := make(chan struct{})
	close(opened)
	return opened
}
func (output *mixerOutput) err() error     { return nil }
func (output *mixerOutput) suspend() error { return nil }
func (output *mixerOutput) resume() error  { return nil }
