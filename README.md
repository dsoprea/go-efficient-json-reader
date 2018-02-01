## Overview

Go provides two built-in strategies to parse JSON data:

- Parse completely (not scalable).
- Tokenize (very primitive).

In most cases, you just want to access one simple object (dictionary) of keys and scalar values or a list of simple objects. Simultaneously, you may have too many to read into memory.

This project tokenizes the JSON data but then also:

- provides a specific token for object keys
- automatically collects all object keys and values, when those values are scalars, and then returns this as a map following all object-close tokens.

This allows you to efficiently parse JSON from an `io.Reader` while also saving you the time of building objects yourself (unless you want something more complicated, such as actually accessing objects of objects).

**The use of this project would provide an advantage in most situations over the built-in tokenizer unless there could be very large objects in your JSON structure that you expect to be ignoring.**


## Example

```go
import (
    "os"
    "path"
    "fmt"

    "github.com/dsoprea/go-efficient-json-reader"
)

func PrintStream() {
    f, err := os.Open("data1.json")

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
            fmt.Printf("FLOAT64: %v\n", token)
        case int64:
            fmt.Printf("INT64: %v\n", token)
        case jsonreader.SimpleObject:
            fmt.Printf("SIMPLE OBJECT: %v\n", token)
        default:
            fmt.Printf("VALUE: %v\n", token)
        }
    }
}
```
