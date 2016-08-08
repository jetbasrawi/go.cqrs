// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"fmt"
)

// StreamNamer is the interface that stream name delegates should implement.
type StreamNamer interface {
	GetStreamName(string, string) (string, error)
}

// DelegateStreamNamer stores delegates per aggregate type allowing fine grained
// control of stream names for event streams.
type DelegateStreamNamer struct {
	delegates map[string]func(string, string) string
}

// NewDelegateStreamNamer constructs a delegate stream namer
func NewDelegateStreamNamer() *DelegateStreamNamer {
	return &DelegateStreamNamer{
		delegates: make(map[string]func(string, string) string),
	}
}

// RegisterDelegate allows registration of a stream name delegate function for
// the aggregates specified in the variadic aggregates argument.
func (r *DelegateStreamNamer) RegisterDelegate(delegate func(string, string) string, aggregates ...AggregateRoot) error {
	for _, aggregate := range aggregates {
		typeName := typeOf(aggregate)
		if _, ok := r.delegates[typeName]; ok {
			return fmt.Errorf("The stream name delegate for \"%s\" is already registered with the stream namer.",
				typeName)
		}
		r.delegates[typeName] = delegate
	}
	return nil
}

// GetStreamName gets the result of the stream name delgate registered for the aggregate type.
func (r *DelegateStreamNamer) GetStreamName(aggregateTypeName string, id string) (string, error) {
	if f, ok := r.delegates[aggregateTypeName]; ok {
		return f(aggregateTypeName, id), nil
	}
	return "", fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
		aggregateTypeName)
}
