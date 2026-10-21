//go:build nintendo3ds

package sound

/*
#include <3ds.h>
#include <stdlib.h>
#include <string.h>

// The DSP plays one channel of 16-bit stereo frames, from buffers taken in turn: while it plays
// one, the others are filled. Buffers live in linear memory, which the DSP reads.
#define SOUND_BUFFERS 6 // soundBuffers, below

static ndspWaveBuf waveBuffers[SOUND_BUFFERS];
static s16* soundData;

static void soundClose(void) {
	ndspExit();
	if (soundData != NULL) {
		linearFree(soundData);
		soundData = NULL;
	}
}

// soundOpen starts the DSP at a rate, with buffers of so many frames; 0 when it cannot (the DSP's
// firmware, /3ds/dspfirm.cdc on the SD card, is missing).
static int soundOpen(int rate, int frames) {
	if (R_FAILED(ndspInit())) return 0;
	soundData = linearAlloc(SOUND_BUFFERS * frames * 2 * sizeof(s16));
	if (soundData == NULL) {
		ndspExit();
		return 0;
	}
	ndspSetOutputMode(NDSP_OUTPUT_STEREO);
	ndspChnReset(0);
	ndspChnSetInterp(0, NDSP_INTERP_LINEAR);
	ndspChnSetRate(0, (float)rate);
	ndspChnSetFormat(0, NDSP_FORMAT_STEREO_PCM16);
	float mix[12];
	memset(mix, 0, sizeof(mix));
	mix[0] = 1.0f; // front left
	mix[1] = 1.0f; // front right
	ndspChnSetMix(0, mix);
	memset(waveBuffers, 0, sizeof(waveBuffers));
	for (int index = 0; index < SOUND_BUFFERS; index++) {
		waveBuffers[index].data_vaddr = soundData + index * frames * 2;
		waveBuffers[index].nsamples = frames;
		waveBuffers[index].status = NDSP_WBUF_DONE;
	}
	atexit(soundClose);
	return 1;
}

static s16* soundBuffer(int index) { return waveBuffers[index].data_pcm16; }

static int soundBufferFree(int index) {
	u8 status = waveBuffers[index].status;
	return status == NDSP_WBUF_DONE || status == NDSP_WBUF_FREE;
}

static void soundQueue(int index, int frames) {
	DSP_FlushDataCache(waveBuffers[index].data_pcm16, frames * 2 * sizeof(s16));
	ndspChnWaveBufAdd(0, &waveBuffers[index]);
}

static void soundPause(int paused) { ndspChnSetPaused(0, paused != 0); }
*/
import "C"

import (
	"fmt"
	"io"
	"time"
	"unsafe"
)

// On the Nintendo 3DS the context mixes its players itself (mixer) and hands the DSP buffers of
// 16-bit stereo frames, through libctru's ndsp. TinyGo's scheduler there is cooperative: the
// buffers are filled from a goroutine, which runs when the program's loop yields, as go-ui's window
// does every frame (a loop of its own calls runtime.Gosched once a frame).

// soundBuffers is how many buffers the DSP takes in turn, and bufferTime how long each plays: a
// fifth of a second ahead in all, room for a few frames that take long (a game loading a ship,
// a network handshake) before the sound crackles.
const (
	soundBuffers = 6
	bufferTime   = time.Second / 30
)

// noDSP is the console's ErrOutputUnavailable: its DSP cannot start without its firmware,
// /3ds/dspfirm.cdc, on the SD card (homebrew tools dump it there).
var noDSP = fmt.Errorf("%w: the DSP cannot start: is /3ds/dspfirm.cdc on the SD card?", ErrOutputUnavailable)

type consoleOutput struct {
	mixer
	frames  int
	buffers [soundBuffers][]int16
	queued  [soundBuffers]bool
	opened  chan struct{}
	failure error
}

func openSystemOutput(rate int, latency time.Duration, name string) (output, error) {
	frames := int(int64(rate) * int64(bufferTime) / int64(time.Second))
	if latency > 0 {
		frames = max(int(int64(rate)*int64(latency)/int64(time.Second))/soundBuffers, 256)
	}
	output := &consoleOutput{frames: frames, opened: make(chan struct{})}
	close(output.opened)
	if C.soundOpen(C.int(rate), C.int(frames)) == 0 {
		output.failure = noDSP
		return output, nil // a program without sound still runs; Err tells why it is silent
	}
	for index := range output.buffers {
		output.buffers[index] = unsafe.Slice((*int16)(unsafe.Pointer(C.soundBuffer(C.int(index)))), 2*frames)
	}
	go output.pump()
	return output, nil
}

// pump fills every buffer the DSP is done with, and counts what was heard.
func (output *consoleOutput) pump() {
	for {
		for index := range output.buffers {
			if C.soundBufferFree(C.int(index)) == 0 {
				continue
			}
			if output.queued[index] {
				output.played(int64(output.frames))
			}
			output.mixInto(output.buffers[index])
			C.soundQueue(C.int(index), C.int(output.frames))
			output.queued[index] = true
		}
		time.Sleep(bufferTime / 4)
	}
}

func (output *consoleOutput) newVoice(source io.ReadSeeker) voice {
	return output.mixer.newVoice(source)
}
func (output *consoleOutput) ready() <-chan struct{} { return output.opened }
func (output *consoleOutput) err() error             { return output.failure }

func (output *consoleOutput) suspend() error {
	if output.failure == nil {
		C.soundPause(1)
	}
	return nil
}

func (output *consoleOutput) resume() error {
	if output.failure == nil {
		C.soundPause(0)
	}
	return nil
}
