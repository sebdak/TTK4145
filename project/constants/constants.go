package constants

type elevatorState int

const (
	INITIALIZING elevatorState = 1 + iota
	AT_FLOOR
	MOVING
)
