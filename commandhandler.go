// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package ycq

import "fmt"

//CommandHandler is the interface that all command handlers should implement.
type CommandHandler interface {
	Handle(CommandMessage) error
}

//CommandHandlerBase is an embedded type that supports chaining of command handlers
//through provision of a next field that will hold a reference to the next handler
//in the chain.
type CommandHandlerBase struct {
	next CommandHandler
}

//CommandExecutionError is the error returned in response to a failed command.
type CommandExecutionError struct {
	Command CommandMessage
	Reason  string
}

func (this *CommandExecutionError) Error() string {
	return fmt.Sprintf("Invalid Operation. Command: %s Reason: %s", this.Command.CommandType(), this.Reason)
}

