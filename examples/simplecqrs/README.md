# Simple CQRS

Simple CQRS is a Golang implementation of the cononical CQRS example application written by Greg Young.

The original C# can be found at https://github.com/gregoryyoung/m-r.

The example has tried to remain as true to the original application as possible to aid learning. The 
original is written in C# and so some of the conventions ported here may not be strictly idiomatic Go, 
however I think the result is easy to read and understand. 

I have tried to strike a balance between remaining true to the original implementation to aid understanding 
and following Golang idioms. 

## How to use

```
$ go run main.go
```

The project will serve a web application on http://localhost:8088 which allows you to add and modify items in an inventory.


