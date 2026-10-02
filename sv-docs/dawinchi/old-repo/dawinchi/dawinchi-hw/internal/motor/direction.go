package motor

type Direction int

const (
	CW  Direction = iota // clockwise, forward
	CCW                  // counter clockwise, backward
)

func (d Direction) String() string {
	switch d {
	case CW:
		return "CW"
	case CCW:
		return "CCW"
	default:
		return "invalid direction"
	}
}
