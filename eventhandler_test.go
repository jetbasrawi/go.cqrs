package ycq

func NewMockEventHandler() *MockEventHandler {
	return &MockEventHandler{
		make([]EventMessage, 0),
	}
}

type MockEventHandler struct {
	events []EventMessage
}

func (m *MockEventHandler) Handle(event EventMessage) {
	m.events = append(m.events, event)
}
