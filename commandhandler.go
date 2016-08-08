// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

// CommandHandler is the interface that all command handlers should implement.
type CommandHandler interface {
	Handle(CommandMessage) error
}

// CommandHandlerBase is an embedded type that supports chaining of command handlers
// through provision of a next field that will hold a reference to the next handler
// in the chain.
type CommandHandlerBase struct {
	next CommandHandler
}
