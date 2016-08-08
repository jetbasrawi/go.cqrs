// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package ycq provides a CQRS reference implementation.
//
// The implementation follows as much as possible the classic reference implementation
// m-r by Greg Young.
//
// The implmentation differs in a number of respects becasue the original is written
// in C# and uses Generics where generics are not available in Go.
// This implementation instead uses interfaces to deal with types in a generic manner
// and used delegate functions to instantiate specific types.
package ycq

import (
	"reflect"

	"github.com/jetbasrawi/go.cqrs/internal/uuid"
)

// typeOf is a convenience function that returns the name of a type
//
// This is used so commonly throughout the code that it is better to
// have this convenience function and also allows for changing the scheme
// used for the type name more easily if desired.
func typeOf(i interface{}) string {
	return reflect.TypeOf(i).Elem().Name()
}

// NewUUID returns a new v4 uuid as a string
func NewUUID() string {
	return uuid.NewUUID()
}
