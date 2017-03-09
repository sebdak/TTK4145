package constants

const NumberOfFloors = 4
const NumberOfElevators = 3


type ElevatorState int
type ElevatorDirection int
type ElevatorButton int

const (
	NewExternalOrderPort = 25069 +iota
	HandledExternalOrderPort 
	HeadingPort
	QueuePort
	PeersPort 
	MasterPort 
)

const (
	Initializing ElevatorState = 1 + iota
	AtFloor
	Moving
	Broken
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

type Order struct {
	Floor      int
	Direction  ElevatorDirection
	ElevatorID string
}

type ElevatorHeading struct {
	LastFloor int
	CurrentOrder Order
	Id string
}