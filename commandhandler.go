package ycq

type CommandHandler interface {
	Handle(CommandMessage) error
}

type CommandHandlerBase struct {
	next CommandHandler
}
