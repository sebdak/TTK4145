package driver
/*
#cgo CFLAGS: -std=c99 -Ilib 
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
#include "channels.h"
#include "elevIO.h"
*/
import "C"
import constants "../constants"

type Elev_button_type_t int

func InitElev() {
	C.elev_init()
}

func SetMotorDir(dir constants.ElevatorDirection) {
	C.elev_set_motor_direction(C.elev_motor_direction_t(dir))
}

func GetFloorSensor() int {
	return int(C.elev_get_floor_sensor_signal())
}

func GetButtonSignal(button Elev_button_type_t, floor int) int {
	return int(C.elev_get_button_signal(C.elev_button_type_t(button), C.int(floor)))
}

func GetStopSignal() int {
	return int(C.elev_get_stop_signal())
}

func GetObstructionSignal() int {
	return int(C.elev_get_stop_signal())
}

func SetFloorIndicator(floor int) {
	C.elev_set_floor_indicator(C.int(floor))
}

func SetButtonLamp(button Elev_button_type_t, floor int, value int) {
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor), C.int(value))
}

func SetStopLamp(value int) {
	C.elev_set_stop_lamp(C.int(value))
}

func SetDoorOpenLamp(value int) {
	C.elev_set_door_open_lamp(C.int(value))
}
