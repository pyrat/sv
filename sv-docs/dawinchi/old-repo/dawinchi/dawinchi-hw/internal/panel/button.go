package panel

// 30	GPIO06	HOME button
// 55	GPIO14	UP button
// 51	GPIO15	DOWN button
// 29	GPIO16	IN button
// 50	GPIO17	OUT button

const (
	ButtonHomePin = 6
	ButtonUpPin   = 14
	ButtonDownPin = 15
	ButtonInPin   = 16
	ButtonOutPin  = 17
)

type ButtonEvent int

const (
	ButtonHomePush ButtonEvent = iota
	ButtonHomeRelease
	ButtonUpPush
	ButtonUpRelease
	ButtonDownPush
	ButtonDownRelease
	ButtonInPush
	ButtonInRelease
	ButtonOutPush
	ButtonOutRelease
)

func (be ButtonEvent) String() string {
	switch be {
	case ButtonHomePush:
		return "ButtonHomePush"
	case ButtonHomeRelease:
		return "ButtonHomeRelease"
	case ButtonUpPush:
		return "ButtonUpPush"
	case ButtonUpRelease:
		return "ButtonUpRelease"
	case ButtonDownPush:
		return "ButtonDownPush"
	case ButtonDownRelease:
		return "ButtonDownRelease"
	case ButtonInPush:
		return "ButtonInPush"
	case ButtonInRelease:
		return "ButtonInRelease"
	case ButtonOutPush:
		return "ButtonOutPush"
	case ButtonOutRelease:
		return "ButtonOutRelease"
	default:
		return "Unknown"
	}
}
