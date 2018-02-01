## Overview

Go provides two built-in strategies to parse JSON data:

- Parse completely (not scalable)
- Tokenize (very primitive)

In most cases, you just want to access one simple object (dictionary) of keys and scalar values or a list of simple objects. Simultaneously, you may have too many to read contiguously into memory.

This project tokenizes the JSON data but then also:

- Provides a specific token for object keys
- Automatically collects all object keys and values, when those values are scalars, and then returns this as a map following all object-close tokens.

This allows you to efficiently parse JSON from an `io.Reader` while also saving you the time of building objects yourself (unless you want something more complicated, such as having direct access to objects that embed other objects).

**The use of this project would provide an advantage in most situations over the built-in tokenizer unless there could be very large objects in your JSON structure that you expect to be ignoring.**


## Example

```go
package main

import (
    "os"
    "fmt"

    "github.com/dsoprea/go-efficient-json-reader"
)

func main() {
    f, err := os.Open("data.json")

    defer f.Close()

    c := make(chan interface{}, 0)

    err = jsonreader.Parse(f, c)
    if err != nil {
        panic(err)
    }

    for token := range c {
        switch token.(type) {
        case jsonreader.ObjectOpen:
            fmt.Printf("OBJECT-OPEN\n")
        case jsonreader.ObjectClose:
            fmt.Printf("OBJECT-CLOSE\n")
        case jsonreader.ListOpen:
            fmt.Printf("LIST-OPEN\n")
        case jsonreader.ListClose:
            fmt.Printf("LIST-CLOSE\n")
        case jsonreader.ObjectKey:
            fmt.Printf("OBJECT-KEY: %s\n", token)
        case float64:
            fmt.Printf("FLOAT64: %f\n", token)
        case int64:
            fmt.Printf("INT64: %d\n", token)
        case string:
            fmt.Printf("STRING: %s\n", token)
        case jsonreader.SimpleObject:
            fmt.Printf("SIMPLE OBJECT: %v\n", token)
        }
    }
}
```

Example data:

```
[
  99,
  123.45,
  [
    678.9
  ],
  {
    "aa": "bb",
    "cc": 1000.1,
    "dd": {
        "subkey1": "subvalue1"
    }
  },
  "test string"
]
```

Output:

```
LIST-OPEN
FLOAT64: 99.000000
FLOAT64: 123.450000
LIST-OPEN
FLOAT64: 678.900000
LIST-CLOSE
OBJECT-OPEN
OBJECT-KEY: aa
STRING: bb
OBJECT-KEY: cc
FLOAT64: 1000.100000
OBJECT-KEY: dd
OBJECT-OPEN
OBJECT-KEY: subkey1
STRING: subvalue1
OBJECT-CLOSE
SIMPLE OBJECT: map[subkey1:subvalue1]
OBJECT-CLOSE
SIMPLE OBJECT: map[aa:bb cc:1000.1]
STRING: test string
LIST-CLOSE
```

Note that it is Go that is parsing all numbers as *float64*s, not this project.
