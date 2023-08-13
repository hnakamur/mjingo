package mjingo

type captureMode uint8

const (
	captureModeCapture captureMode = iota + 1
	captureModeDiscard
)
