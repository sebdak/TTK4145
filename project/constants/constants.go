package constants

type ElevatorState int
type ElevatorDirection int

const (
	INITIALIZING ElevatorState = 1 + iota
	AT_FLOOR
	MOVING
)

const (
	DIR_DOWN ElevatorDirection = -1
	DIR_STOP = 0
	DIR_UP = 1
)
