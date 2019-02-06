# Go.CQRS [![license](https://img.shields.io/badge/license-MIT-blue.svg?maxAge=2592000)](https://github.com/jetbasrawi/go.cqrs/blob/master/LICENSE.md) [![Go Report Card](https://goreportcard.com/badge/github.com/jetbasrawi/go.cqrs)](https://goreportcard.com/report/github.com/jetbasrawi/go.cqrs) [![GoDoc](https://godoc.org/github.com/jetbasrawi/go.cqrs?status.svg)](https://godoc.org/github.com/jetbasrawi/go.cqrs)


## A Golang CQRS Reference implementation

Go.CQRS provides interfaces and implementations to support a CQRS implementation in Golang. The examples 
directory contains a sample application that demonstrates how to use Go.CQRS.

As much as possible Go.CQRS has been designed with the principles of CQRS espoused by Greg Young which 
represents the best thinking on the topic.

## CQRS Pattern vs CQRS Framework

CQRS is an architectural pattern. When implementing the CQRS pattern, it is easy to imagine how the code 
could be packaged into a framework. However, it is recommended that those working with CQRS focus on learning
the underlying detail of the pattern rather than simply use a framework.

The implementation of the CQRS pattern is not especially difficult, however it is a steep learning curve because 
the pattern is very different to the traditional non CQRS architecture. Topics such as Aggregate Design are very 
different. If you are going to use EventSourcing and eventual consistency then there is a lot of learning to be 
done.

If you are new to CQRS or simply interested in best practices there is a great 6 hour video of a 
[hands-on CQRS](https://www.youtube.com/watch?v=whCk1Q87_ZI) workshop by Greg Young.

Once the pattern is understood, implementations such as Go.CQRS can be used as a reference for learning how to 
implement the pattern in Golang and also as a foundation upon which to build your CQRS implementation.

## What does Go.CQRS provide?

|Feature|Description|
|-------|-----------|
| **Aggregate** | AggregateRoot interface and Aggregate base type that can be embedded in your own types to provide common functions required by aggregates |
| **Event** | An Event interface and an EventDescriptor which is a message envelope for events. Events in Go.CQRS are simply plain Go structs and there are no magic strings to describe them as is the case in some other Go implementations. |
| **Command** | A Command interface and an CommandDescriptor which is a message envelope for commands. Commands in Go.CQRS are simply plain Go structs and there are no magic strings to describe them as is the case in some other Go implementations. | 
| **CommandHandler**| Interface and base functionality for chaining command handlers |
| **Dispatcher** | Dispatcher interface and an in memory dispatcher implementation |
| **EventBus** | EventBus interface and in memory implementation |
| **EventHandler** | EventHandler interface |
| **Repository** | Repository interface and an implementation of the CommonDomain repository that persists events in [GetEventStore](https://geteventstore.com/). While there are many generic event store implementations over common databases such as MongoDB,   [GetEventStore](https://geteventstore.com/) is a specialised EventSourcing database that is open source, performant and reflects the best thinking on the topic from a highly experienced team in this field. |
| **StreamNamer** | A StreamNamer interface and a DelegateStreamNamer implementation that supports the use of functions with the signiature **func(string, string) string** to provide flexibility around stream naming. A common way to construct a stream name might be to use the name of your **BoundedContext** suffixed with an AggregateID. | 

All implementations are easily replaced to suit your particular requirements.

## Example code
The examples folder contains a simple and clear example of how to use go.cqrs to contruct your service. The example is a port of the classic reference implementation [m-r](https://github.com/gregoryyoung/m-r) by [Greg Young](https://github.com/gregoryyoung).

## Getting Started

```
    $ go get github.com/jetbasrawi/go.cqrs

```

Refer to the example application for guidance on how to use Go.CQRS.
