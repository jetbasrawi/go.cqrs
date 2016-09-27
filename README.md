#Go.CQRS [![license](https://img.shields.io/badge/license-MIT-blue.svg?maxAge=2592000)](https://github.com/jetbasrawi/go.cqrs/blob/master/LICENSE.md) [![Go Report Card](https://goreportcard.com/badge/github.com/jetbasrawi/go.cqrs)](https://goreportcard.com/report/github.com/jetbasrawi/go.cqrs) [![GoDoc](https://godoc.org/github.com/jetbasrawi/go.cqrs?status.svg)](https://godoc.org/github.com/jetbasrawi/go.cqrs)


##A Golang CQRS Reference implementation

Go.CQRS provides a classic reference implementation of the CQRS pattern with EventSourcing. 

The implementation provided here comes with a CommonDomainRepository interface implementation that persists 
events in [GetEventStore](https://geteventstore.com/).

##Features
- Plain Go sructs for Events and Commands. No magic strings.
- Common domain repository implementation using [GetEventStore](https://geteventstore.com/) to store events.
- All components defined by interface so the are easily pluggable.

##Getting Started

```
    $ go get github.com/jetbasrawi/go.cqrs

```
The examples folder contains a simple and clear example of how to use go.cqrs.

As much as possible, the application replicates the classic reference 
implementation [m-r](https://github.com/gregoryyoung/m-r) by [Greg Young](https://github.com/gregoryyoung).

