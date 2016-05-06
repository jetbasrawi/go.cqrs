package ycq

type EventHandler interface {
	Handle(EventMessage)
}
