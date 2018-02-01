package jsonreader

import (
    "io"

    "encoding/json"

    "github.com/dsoprea/go-logging"
)

type ObjectOpen rune
type ObjectClose rune
type ListOpen rune
type ListClose rune

type ObjectKey string
type Value interface{}

type SimpleObject map[string]interface{}

func Parse(r io.Reader, c chan<- interface{}) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

    go func() {
        d := json.NewDecoder(r)

        delimiterStack := []rune {}
        previousWasKey := false
        thisWasKey := false
        simpleObjectStack := []map[string]interface{} {}
        previousKey := ""

        for {
            t, err := d.Token()
            if err != nil {
                if err == io.EOF {
                    break
                }

                log.PanicIf(err)
            }

            var lastObject map[string]interface{}

            switch t.(type) {
            case json.Delim:
                r := rune(t.(json.Delim))

                if r == '{' {
                    // Entering an object.

                    delimiterStack = append(delimiterStack, r)

                    c <- ObjectOpen(r)

                    // Create an instance to add any keys having scalar values.
                    simpleObjectStack = append(simpleObjectStack, make(map[string]interface{}))
                } else if r == '}' {
                    // Leaving an object.

                    if len(delimiterStack) == 0 {
                        log.Panicf("too many closing curly-brackets (at root)")
                    }

                    len_ := len(delimiterStack)
                    var last json.Token
                    last, delimiterStack = delimiterStack[len_ - 1], delimiterStack[:len_ - 1]

                    if last != '{' {
                        log.Panicf("too many closing curly-brackets (inside)")
                    }

                    c <- ObjectClose(r)

                    // Also, feed a whole object that we've added any keys and
                    // scalar values that we've encountered to.

                    len_ = len(simpleObjectStack)
                    lastObject, simpleObjectStack = simpleObjectStack[len_ - 1], simpleObjectStack[:len_ - 1]
                    c <- SimpleObject(lastObject)
                } else if r == '[' {
                    // Entering a list.

                    delimiterStack = append(delimiterStack, r)

                    c <- ListOpen(r)
                } else if r == ']' {
                    // Leaving a list.

                    if len(delimiterStack) == 0 {
                        log.Panicf("too many closing square-brackets (at root)")
                    }

                    len_ := len(delimiterStack)
                    var last json.Token
                    last, delimiterStack = delimiterStack[len_ - 1], delimiterStack[:len_ - 1]

                    if last != '[' {
                        log.Panicf("too many closing square-brackets (inside)")
                    }

                    c <- ListClose(r)
                }

            case bool:
                c <- Value(t.(bool))

                // If we're processing the value for a key, set the pair into
                // the last simple object that we created.
                if previousKey != "" {
                    len_ := len(simpleObjectStack)
                    lastObject = simpleObjectStack[len_ - 1]
                    lastObject[previousKey] = t.(bool)
                }
            case float64:
                c <- Value(t.(float64))

                // If we're processing the value for a key, set the pair into
                // the last simple object that we created.
                if previousKey != "" {
                    len_ := len(simpleObjectStack)
                    lastObject = simpleObjectStack[len_ - 1]
                    lastObject[previousKey] = t.(float64)
                }
            case string:
                if previousWasKey == false && len(delimiterStack) > 0 && delimiterStack[len(delimiterStack) - 1] == '{' {
                    c <- ObjectKey(t.(string))

                    // Keep track of how we're encountering keysso that we can
                    // identify the values.
                    thisWasKey = true
                    previousKey = t.(string)
                } else {
                    c <- Value(t.(string))

                    // If we're processing the value for a key, set the pair
                    // into the last simple object that we created.
                    if previousKey != "" {
                        len_ := len(simpleObjectStack)
                        lastObject = simpleObjectStack[len_ - 1]
                        lastObject[previousKey] = t.(string)
                    }
                }

            case nil:
            }

            // The ordering is important here, in the event that the value is
            // another object.

            // If we've finished processing an object value.
            if previousWasKey == true {
                previousWasKey = false
                previousKey = ""
            }

            // If we're finished processing an object key.
            if thisWasKey == true {
                previousWasKey = true
                thisWasKey = false
            }
        }

        close(c)
    }()

    return err
}

func ParseToTokenSlice(r io.Reader) (ts []interface{}, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

    c := make(chan interface{}, 0)

    err = Parse(r, c)
    log.PanicIf(err)

    ts = make([]interface{}, 0)
    for token := range c {
        ts = append(ts, token)
    }

    return ts, nil
}
