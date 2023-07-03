redjet is a high-performance Go library for Redis. Its hallmark feature is
a low-allocation, streaming API.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Introduction](#introduction)
- [Basic Usage](#basic-usage)
- [Streaming](#streaming)
- [Pipelining](#pipelining)
- [Benchmarks](#benchmarks)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Introduction

Unlike redigo and go-redis, redjet does not provide a function for every
Redis command. Instead, it offers a generic interface that supports [all commands
and options](https://redis.io/commands/). While this approach has less
type-safety, it provides forward compatibility with new Redis features.
# Basic Usage

For the most part, you can interact with Redis using a familiar interface:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ammario/redjet"
)

func main() {
    client := redjet.New("localhost:6379")
    ctx := context.Background()

    err := client.Command(ctx, "SET", "foo", "bar").Ok()
    // check error

    got, err := client.Command(ctx, "GET", "foo").Bytes()
    // check error
    // got == []byte("bar")
}
```

# Streaming

When it comes time for performance, you may call `WriteTo` on the result
instead of `Bytes`, which will stream the response directly to an `io.Writer` such as a file or HTTP response.

Similarly, you can pass in a value that implements `redjet.LenReader` to
`Command` to stream larger readers into Redis.

# Pipelining

`redjet` supports pipelining via the `Pipeline` method. This method accepts a Result, potentially that of a previous command.

```go
// Set foo0, foo1, ..., foo99 to "bar", and confirm that each succeeded.
//
// This entire example only takes one round-trip to Redis!
var r *Result
for i := 0; i < 100; i++ {
    r = client.Pipeline(r, "SET", fmt.Sprintf("foo%d", i), "bar")
}

for r.Next() {
    if err := r.Ok(); err != nil {
        log.Fatal(err)
    }
}
```

# Benchmarks

On a pure throughput basis, redjet will perform similarly to redigo and go-redis.
But, since redjet doesn't allocate memory for the entire response object, it
consumes far less resources when handling large responses.

Here are some benchmarks (reproducible via `make gen-bench`) to illustrate:

```
goos: darwin
goarch: arm64
pkg: github.com/ammario/redjet/bench
 │   Redjet    │               Redigo               │              GoRedis               │
 │   sec/op    │   sec/op     vs base               │   sec/op     vs base               │
   1.287m ± 4%   1.374m ± 1%  +6.81% (p=0.000 n=10)   1.379m ± 4%  +7.21% (p=0.000 n=10)

 │    Redjet    │               Redigo                │               GoRedis               │
 │     B/s      │     B/s       vs base               │     B/s       vs base               │
   777.2Mi ± 4%   727.7Mi ± 1%  -6.37% (p=0.000 n=10)   724.9Mi ± 4%  -6.72% (p=0.000 n=10)

 │   Redjet    │                    Redigo                    │                   GoRedis                    │
 │    B/op     │      B/op        vs base                     │      B/op        vs base                     │
   66.00 ± 12%   1047441.50 ± 0%  +1586932.58% (p=0.000 n=10)   1057013.50 ± 0%  +1601435.61% (p=0.000 n=10)

 │   Redjet   │               Redigo                │              GoRedis               │
 │ allocs/op  │  allocs/op   vs base                │ allocs/op   vs base                │
   4.000 ± 0%   2.000 ± 50%  -50.00% (p=0.000 n=10)   6.000 ± 0%  +50.00% (p=0.000 n=10)
```


Note that they are a bit contrived in that they Get a 1MB object. The performance
of all libraries converge as response size decreases. If you don't
need the performance this library offers, you should probably use a more
well-tested library like redigo or go-redis.