//go:build !nintendo3ds

package sound

import (
	"io"
	"time"

	"github.com/ebitengine/oto/v3"
)

// systemOutput sends the sound to the system through oto, which mixes the voices.
type systemOutput struct {
	context *oto.Context
	opened  chan struct{}
}

func openSystemOutput(rate int, latency time.Duration, name string) (output, error) {
	context, opened, err := oto.NewContext(&oto.NewContextOptions{SampleRate: rate, ChannelCount: 2,
		Format: oto.FormatFloat32LE, BufferSize: latency, ApplicationName: name})
	if err != nil {
		return nil, outputUnavailable(err.Error())
	}
	return &systemOutput{context: context, opened: opened}, nil
}

func (system *systemOutput) newVoice(source io.ReadSeeker) voice {
	return system.context.NewPlayer(source)
}
func (system *systemOutput) ready() <-chan struct{} { return system.opened }
func (system *systemOutput) err() error             { return system.context.Err() }
func (system *systemOutput) suspend() error         { return system.context.Suspend() }
func (system *systemOutput) resume() error          { return system.context.Resume() }
