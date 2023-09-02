package mjingo

import "fmt"

type span struct {
	StartLine   uint32
	StartCol    uint32
	StartOffset uint32
	EndLine     uint32
	EndCol      uint32
	EndOffset   uint32
}

func (s *span) String() string {
	return fmt.Sprintf(" @ %d:%d-%d:%d", s.StartLine, s.StartCol, s.EndLine, s.EndCol)
}
