package ycq

import (
	"fmt"
	"reflect"

	"github.com/jetbasrawi/yoono-uuid"
)

type CommandExecutionError struct {
	Command CommandMessage
	Reason  string
}

func (this *CommandExecutionError) Error() string {
	return fmt.Sprintf("Invalid Operation. Command: %s Reason: %s", this.Command.CommandType(), this.Reason)
}

func typeOf(i interface{}) string {
	return reflect.TypeOf(i).Elem().Name()
}

// A helper function to provide quick access to a uuid
// sorry about the naming. I could not resist the pun
func yooid() uuid.UUID {
	return uuid.NewV4()
}
