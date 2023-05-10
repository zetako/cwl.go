package runner

// Signal to guide engine process
type Signal string

const (
	SignalPause  Signal = "pause"
	SignalResume Signal = "resume"
	SignalAbort  Signal = "abort"
)
