package constants

const NumberOfFloors = 4
const NumberOfElevators = 1
const MessagePort = 20070
const PeersPort = MessagePort+1337


type ElevatorState int
type ElevatorDirection int
type ElevatorButton int

const (
	Initializing ElevatorState = 1 + iota
	AtFloor
	Moving
)

const (
	DirUp   ElevatorDirection = 1
	DirStop                   = 0
	DirDown                   = -1
)

const (
	ButtonCallUp ElevatorButton = 0 + iota
	ButtonCallDown
	ButtonCommand
)

type NewOrder struct {
	Floor      int
	Direction  ElevatorDirection
	ElevatorID int
}
