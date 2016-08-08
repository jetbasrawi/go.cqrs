#Go.CQRS [![license](https://img.shields.io/badge/license-MIT-blue.svg?maxAge=2592000)](https://github.com/jetbasrawi/go.cqrs/blob/master/LICENSE.md) [![Go Report Card](https://goreportcard.com/badge/github.com/jetbasrawi/go.cqrs)](https://goreportcard.com/report/github.com/jetbasrawi/go.cqrs) [![GoDoc](https://godoc.org/github.com/jetbasrawi/go.cqrs?status.svg)](https://godoc.org/github.com/jetbasrawi/go.cqrs)


##A Golang CQRS Reference implementation

Go.CQRS provides a simple Golang reference implementation of the CQRS pattern. 

As much as possible, the implementation has tried to replicate the classic reference 
implementation [m-r](https://github.com/gregoryyoung/m-r) by [Greg Young](https://github.com/gregoryyoung).

The implementation provided here comes with a CommonDomainRepository interface implementation that persists 
events in [GetEventStore](https://geteventstore.com/).

```
    $ go get github.com/jetbasrawi/go.cqrs

```

