package record

import (
	"time"

	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/logger"
	"github.com/bluenviron/mediamtx/internal/storage"
	"github.com/bluenviron/mediamtx/internal/stream"
	RMS "github.com/bluenviron/mediamtx/internal/grps"
)

// OnSegmentCreateFunc is the prototype of the function passed as OnSegmentCreate
type OnSegmentCreateFunc = func(path string)

// OnSegmentCompleteFunc is the prototype of the function passed as OnSegmentComplete
type OnSegmentCompleteFunc = func(path string, duration time.Duration)

// Agent writes recordings to disk.
type Agent struct {
	WriteQueueSize    int
	PathFormat        string
	PathFormats       map[string]string
	Format            conf.RecordFormat
	PartDuration      time.Duration
	SegmentDuration   time.Duration
	PathName          string
	StreamName        string
	Stream            *stream.Stream
	OnSegmentCreate   OnSegmentCreateFunc
	OnSegmentComplete OnSegmentCompleteFunc
	Parent            logger.Writer

	restartPause time.Duration

	currentInstance *agentInstance

	terminate chan struct{}
	done      chan struct{}

	ClientGRPC      RMS.GrpcClient
	Stor        storage.Storage
	RecordAudio bool

	PathStream    string
	CodeMp        string
	Status_record int8


	Pathrecord  bool
	ChConfigSet chan []struct {
		Name   string
		Record bool
	}
}

// Initialize initializes Agent.
func (w *Agent) Initialize() {
	if w.OnSegmentCreate == nil {
		w.OnSegmentCreate = func(string) {
		}
	}
	if w.OnSegmentComplete == nil {
		w.OnSegmentComplete = func(string, time.Duration) {
		}
	}
	if w.restartPause == 0 {
		w.restartPause = 2 * time.Second
	}

	w.terminate = make(chan struct{})
	w.done = make(chan struct{})

	w.currentInstance = &agentInstance{
		agent:       w,
		stor:        w.Stor,
		clientGRPC:  w.ClientGRPC,
		recordAudio: w.RecordAudio,
	}
	w.currentInstance.initialize()

	go w.run()
}

// Log implements logger.Writer.
func (w *Agent) Log(level logger.Level, format string, args ...interface{}) {
	w.Parent.Log(level, "[record] "+format, args...)
}

// Close closes the agent.
func (w *Agent) Close() {
	w.Log(logger.Info, "recording stopped")
	close(w.terminate)
	<-w.done
}

func (w *Agent) run() {
	defer close(w.done)

	for {
		select {
		case <-w.currentInstance.done:
			w.currentInstance.close()
		case <-w.terminate:
			w.currentInstance.close()
			return
		}

		select {
		case <-time.After(w.restartPause):
		case <-w.terminate:
			return
		}

		w.currentInstance = &agentInstance{
			agent:       w,
			stor:        w.Stor,
			recordAudio: w.RecordAudio,
		}
		w.currentInstance.initialize()
	}
}
