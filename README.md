# gocovr [![Build Status](https://travis-ci.org/thatguystone/gocovr.svg)](https://travis-ci.org/thatguystone/gocovr)

Very simple coverage reporting for golang, that supports coverage reporting on multiple packages!

Sample output:

```
$ gocovr test

Base: github.com/example/repo

File   Lines  Exec  Cover   Missing

a.go   66     59    89.4%   79-82,96-97,152-154,187-197
b.go   23     17    73.9%   37-42,63-66,70-72
c.go   23     18    78.3%   46-48,51-53,56-58,66-72
d.go   21     21    100.0%
e.go   37     0     0.0%    31-108


TOTAL  170    115   67.6%
```

## Installation

```bash
go get github.com/thatguystone/gocovr
```

## Usage

There are currently 2 ways to use gocovr:

1. in place of golang's `go test`
2. a replacement for opening a browser to view coverage

### gocovr test

Running `gocovr test` will run `go test` with an added `-coverprofile=cover.out`. Any arguments following `test` are transparently passed through to `go test`.

### gocovr [cover.out]

Running `gocovr` will interpret a `cover.out` file by default; you may pass in any other file to parse.

## Arguments

* `-parallel=$GOMAXPROCS`: run this many package tests in parallel
* `-filter='.*'`: only display files that match the given regex

## Skipping

Sometimes you have files that you just don't need to test, and you want to skip them. To do this, add the following line, on its own line, anywhere in the file you want to skip:

```go
//gocovr:skip-file
```
