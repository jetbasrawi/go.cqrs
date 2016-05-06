package ycq

import (
	"fmt"
)

type CommandHandler interface {
	Handle(CommandMessage) error
}

type AggregateCommandHandler struct {
	repository DomainRepository
	aggregates map[string]string
}

func NewAggregateCommandHandler(repository DomainRepository) (*AggregateCommandHandler, error) {
	if repository == nil {
		return nil, fmt.Errorf("The command dispatcher requires a repository.")
	}

	h := &AggregateCommandHandler{
		repository: repository,
		aggregates: make(map[string]string),
	}
	return h, nil
}

// In CQRS, there should only ever be one handler for a command.
func (h *AggregateCommandHandler) RegisterCommands(aggregate AggregateRoot, commands ...interface{}) error {
	for _, cmd := range commands {
		if _, ok := h.aggregates[typeOf(cmd)]; ok {
			return fmt.Errorf("The command \"%s\" is already registered with the dispatcher.", typeOf(cmd))
		}
		h.aggregates[typeOf(cmd)] = typeOf(aggregate)
	}
	return nil
}

func (h *AggregateCommandHandler) Handle(command CommandMessage) error {
	aggregateType, ok := h.aggregates[command.CommandType()]
	if !ok {
		return fmt.Errorf("The dispatcher has no handler registered for commands of type: \"%s\"", command.CommandType())
	}

	aggregate, err := h.repository.Load(aggregateType, command.AggregateID())
	if err != nil {
		return err
	}

	if err = aggregate.Handle(command); err != nil {
		return err
	}

	if err = h.repository.Save(aggregate); err != nil {
		return err
	}

	return nil
}
