# gocovr

Very simple coverage reporting for golang.

Sample output:

```
Base: github.com/example/repo

File   Lines  Exec  Cover   Missing

a.go   66     59    89.4%   79-82,96-97,152-154,187-197
b.go   23     17    73.9%   37-42,63-64,64-66,70-72
c.go   23     18    78.3%   46-48,51-53,56-58,66-68,70-72
d.go   21     21    100.0%
e.go   37     0     0.0%    31-39,41-44,44-46,46-47,47-49,50,53,53-55,58-59,63-69,71-76,76-78,81-93,95-96,96-99,101-102,105-108

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

### gocovr

Running `gocovr` will intepret a `cover.out` file.

## Arguments

* `-file=cover.out`: parse the given coverage file
* `-filter='.*'`: only display files that match the given regex
