package constants

const NumberOfFloors = 4
const NumberOfElevators = 3
const OrderPort = 25069
const HeadingPort = OrderPort+1
const QueuePort = OrderPort+2
const PeersPort = OrderPort+3
const MasterPort = OrderPort+4


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