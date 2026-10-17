package sound

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"sync"
)

// mixer mixes voices itself, for an output that takes one stream of samples instead of mixing
// players as oto does: the Nintendo 3DS's DSP is handed buffers of 16-bit stereo frames. It needs
// no system and runs anywhere, so the tests try it as it is.
type mixer struct {
	mutex   sync.Mutex
	voices  []*mixedVoice
	scratch []byte
	sum     []float32
}

// mixedVoice is a voice of a mixer: its source, whether it plays, how loud, and how many of its
// frames were mixed but not heard yet.
type mixedVoice struct {
	mixer   *mixer
	source  io.ReadSeeker
	playing bool
	volume  float64
	waiting int64 // frames mixed into buffers the output has not played yet
	closed  bool
}

func (mixer *mixer) newVoice(source io.ReadSeeker) voice {
	mixer.mutex.Lock()
	defer mixer.mutex.Unlock()
	made := &mixedVoice{mixer: mixer, source: source, volume: 1}
	mixer.voices = append(mixer.voices, made)
	return made
}

// mixInto fills frames of interleaved 16-bit stereo with every playing voice, at its volume. A
// voice that reaches its end stops; silence fills what nobody plays.
func (mixer *mixer) mixInto(frames []int16) {
	mixer.mutex.Lock()
	defer mixer.mutex.Unlock()
	count := len(frames) / 2
	if cap(mixer.sum) < 2*count {
		mixer.sum = make([]float32, 2*count)
		mixer.scratch = make([]byte, count*bytesPerFrame)
	}
	sum := mixer.sum[:2*count]
	clear(sum)
	for _, voice := range mixer.voices {
		if !voice.playing || voice.closed {
			continue
		}
		mixer.add(voice, sum, count)
	}
	for index, sample := range sum {
		frames[index] = int16(min(max(sample, -1), 1) * math.MaxInt16)
	}
}

// add reads a voice's next frames and adds them to the sum.
func (mixer *mixer) add(voice *mixedVoice, sum []float32, count int) {
	bytes := mixer.scratch[:count*bytesPerFrame]
	read := 0
	for read < len(bytes) {
		got, err := voice.source.Read(bytes[read:])
		read += got
		if err != nil {
			if errors.Is(err, io.EOF) {
				voice.playing = false
			}
			break
		}
		if got == 0 {
			break
		}
	}
	volume := float32(voice.volume)
	samples := read / 4
	for index := 0; index < samples; index++ {
		sum[index] += math.Float32frombits(binary.LittleEndian.Uint32(bytes[4*index:])) * volume
	}
	voice.waiting += int64(read / bytesPerFrame)
}

// played says the output has played frames of what was mixed: they are no longer waiting.
func (mixer *mixer) played(frames int64) {
	mixer.mutex.Lock()
	defer mixer.mutex.Unlock()
	for _, voice := range mixer.voices {
		voice.waiting = max(voice.waiting-frames, 0)
	}
	kept := mixer.voices[:0]
	for _, voice := range mixer.voices {
		if !voice.closed {
			kept = append(kept, voice)
		}
	}
	mixer.voices = kept
}

func (voice *mixedVoice) Play() {
	voice.mixer.mutex.Lock()
	defer voice.mixer.mutex.Unlock()
	voice.playing = true
}

func (voice *mixedVoice) Pause() {
	voice.mixer.mutex.Lock()
	defer voice.mixer.mutex.Unlock()
	voice.playing = false
}

func (voice *mixedVoice) IsPlaying() bool {
	voice.mixer.mutex.Lock()
	defer voice.mixer.mutex.Unlock()
	return voice.playing
}

func (voice *mixedVoice) SetVolume(volume float64) {
	voice.mixer.mutex.Lock()
	defer voice.mixer.mutex.Unlock()
	voice.volume = volume
}

// BufferedSize is what was mixed but not heard yet, in the bytes the voice reads.
func (voice *mixedVoice) BufferedSize() int {
	voice.mixer.mutex.Lock()
	defer voice.mixer.mutex.Unlock()
	return int(voice.waiting) * bytesPerFrame
}

func (voice *mixedVoice) Seek(offset int64, whence int) (int64, error) {
	voice.mixer.mutex.Lock()
	defer voice.mixer.mutex.Unlock()
	voice.waiting = 0
	return voice.source.Seek(offset, whence)
}

func (voice *mixedVoice) Close() error {
	voice.mixer.mutex.Lock()
	defer voice.mixer.mutex.Unlock()
	voice.closed, voice.playing = true, false
	return nil
}
