package ycq

import (
	"fmt"
	"github.com/jetbasrawi/yoono-uuid"
)

type StreamNamer interface {
	GetStreamName(string, uuid.UUID) (string, error)
}

type DelegateStreamNamer struct {
	delegates map[string]func(string, uuid.UUID) string
}

func NewDelegateStreamNamer() *DelegateStreamNamer {
	return &DelegateStreamNamer{
		delegates: make(map[string]func(string, uuid.UUID) string),
	}
}

func (r *DelegateStreamNamer) RegisterDelegate(delegate func(string, uuid.UUID) string, aggregates ...AggregateRoot) error {
	for _, aggregate := range aggregates {
		typeName := aggregate.AggregateType()
		if _, ok := r.delegates[typeName]; ok {
			return fmt.Errorf("The stream name delegate for \"%s\" is already registered with the stream namer.",
				typeName)
		}
		r.delegates[typeName] = delegate
	}
	return nil
}

func (r *DelegateStreamNamer) GetStreamName(aggregateTypeName string, id uuid.UUID) (string, error) {
	if f, ok := r.delegates[aggregateTypeName]; ok {
		return f(aggregateTypeName, id), nil
	}
	return "", fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
		aggregateTypeName)
}
