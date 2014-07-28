package delta

import (
	"state"
)

type Delta struct {
	Offset uint16
	Data   []byte
}

type SegmentDiff struct {
	UploadID state.UploadID
	DeltaSet []Delta
}

func (e *Engine) UpdateSegment(sd SegmentDiff) (err error) {
	return
}
