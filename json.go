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

// ObjectValue represents any right-side object token. Note that the switch
// statement can distinguish a string from an ObjectKey correctly but not an
// ObjectValue type-alias of an interface{}. So, we *encapsulate* an
// interface{} rather than aliasing one.
type ObjectValue struct {
    value interface{}
}

func (ov ObjectValue) Value() interface{} {
    return ov.value
}

type Value interface{}

type SimpleObject map[string]interface{}

type Parser struct {
    d *json.Decoder

    delimiterStack []rune
    simpleObjectStack []map[string]interface{}
}

func NewParser(r io.Reader) *Parser {
    d := json.NewDecoder(r)

    return &Parser{
        d: d,

        delimiterStack: make([]rune, 0),
        simpleObjectStack: make([]map[string]interface{}, 0),
    }
}

func (p *Parser) popDelimiter() (r rune, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

    len_ := len(p.delimiterStack)
    if len_ == 0 {
        log.Panicf("unbalanced delimiters")
    }

    var last rune
    last, p.delimiterStack = p.delimiterStack[len_ - 1], p.delimiterStack[:len_ - 1]

    return last, nil
}

func (p *Parser) processDelimiter(c chan<- interface{}, r rune) (ascend bool, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

    if r == '{' {
        // Entering an object.

        p.delimiterStack = append(p.delimiterStack, r)

        c <- ObjectOpen(r)

        // Create an instance to add any keys having scalar values.
        p.simpleObjectStack = append(p.simpleObjectStack, make(map[string]interface{}))

        err = p.parse(c, r)
        log.PanicIf(err)

        return false, nil
    } else if r == '}' {
        // Leaving an object.

        last, err := p.popDelimiter()
        log.PanicIf(err)

        if last != '{' {
            log.Panicf("object closer unbalanced")
        }

        c <- ObjectClose(r)

        // Also, feed a whole object that we've added any keys and
        // scalar values that we've encountered to.

        len_ := len(p.simpleObjectStack)
        var lastObject map[string]interface{}
        lastObject, p.simpleObjectStack = p.simpleObjectStack[len_ - 1], p.simpleObjectStack[:len_ - 1]
        c <- SimpleObject(lastObject)

        return true, nil
    } else if r == '[' {
        // Entering a list.

        p.delimiterStack = append(p.delimiterStack, r)

        c <- ListOpen(r)

        err = p.parse(c, r)
        log.PanicIf(err)

        return false, nil
    } else if r == ']' {
        // Leaving a list.

        last, err := p.popDelimiter()
        log.PanicIf(err)

        if last != '[' {
            log.Panicf("list closer unbalanced")
        }

        c <- ListClose(r)

        return true, nil
    }

    // Should never reach here.
    log.Panic("delimiter processing panic")
    return false, nil
}

func (p *Parser) parse(c chan<- interface{}, startingDelimiter rune) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

// TODO(dustin): !! Implement a second channel and a waitgroup for error-handling from the goroutine.

    previousWasKey := false
    thisWasKey := false
    previousKey := ""
    // var lastToken json.Token

    // Lets us kep track of whether we're on the key or value when processing an
    // object.
    i := 0
    isInObject := startingDelimiter == '{'

    for {
        t, err := p.d.Token()
        if err != nil {
            if err == io.EOF {
                break
            }

            log.PanicIf(err)
        }

        if delimiter, ok := t.(json.Delim); ok == true {
            r := rune(delimiter)

            ascend, err := p.processDelimiter(c, r)
            log.PanicIf(err)

            // We've finished processing an object or a list.
            if ascend == true {
                return nil
            }
        } else {
            isObjectValue := isInObject && i % 2 == 1

            switch t.(type) {
            case bool:
                value := t.(bool)

                // If we're processing the value for a key, set the pair into
                // the last simple object that we created.
                if isObjectValue {
                    len_ := len(p.simpleObjectStack)
                    lastObject := p.simpleObjectStack[len_ - 1]
                    lastObject[previousKey] = value

                    c <- ObjectValue{value: value}
                } else {
                    c <- Value(value)
                }
            case float64:
                value := t.(float64)

                // If we're processing the value for a key, set the pair into
                // the last simple object that we created.
                if isObjectValue {
                    len_ := len(p.simpleObjectStack)
                    lastObject := p.simpleObjectStack[len_ - 1]
                    lastObject[previousKey] = value

                    c <- ObjectValue{value: value}
                } else {
                    c <- Value(value)
                }
            case string:
                value := t.(string)

                if isInObject {
                    if isObjectValue {
                        // We're on an object value.

                        len_ := len(p.simpleObjectStack)
                        lastObject := p.simpleObjectStack[len_ - 1]
                        lastObject[previousKey] = value

                        c <- ObjectValue{value: value}
                    } else if i % 2 == 0 {
                        // We're on an object key.

                        previousKey = value
                        c <- ObjectKey(value)
                    }
                } else {
                    // We're on a string but not in an object (not an object
                    // key, not an object value).

                    c <- Value(value)
                }
            }

            if isInObject {
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

            i++
        }
    }

    return nil
}

func (p *Parser) Parse(c chan<- interface{}) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

    go func() {
        defer func() {
            if state := recover(); state != nil {
                log.Panic(err)
            }
        }()

        err = p.parse(c, 0)
        log.PanicIf(err)

        close(c)
    }()

    return err
}

func (p *Parser) ParseToTokenSlice(r io.Reader) (ts []interface{}, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

    c := make(chan interface{}, 0)

    err = p.Parse(c)
    log.PanicIf(err)

    ts = make([]interface{}, 0)
    for token := range c {
        ts = append(ts, token)
    }

    return ts, nil
}
