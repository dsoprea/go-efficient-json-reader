package jsonreader

import (
    "io"
    // "fmt"
    // "strings"

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
    key string
    value interface{}
}

func (ov ObjectValue) Key() string {
    return ov.key
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

// ObjectContext is relevant if we're processing through an object.
type ObjectContext struct {
    context map[string]interface{}
}

func (oc ObjectContext) Get(key string) interface{} {
    return oc.context[key]
}

// ListContext is relevant if we're processing through a list.
type ListContext struct {
    context map[string]interface{}
}

func (lc ListContext) Get(key string) interface{} {
    return lc.context[key]
}

type Context interface {
    Get(key string) interface{}
}

type delimiterChain struct {
    parent *delimiterChain

    delimiter rune
    context Context

    depth int
}

func (dc delimiterChain) Add(delimiter rune, context Context) delimiterChain {
    return delimiterChain{
        parent: &dc,
        delimiter: delimiter,
        context: context,
        depth: dc.depth + 1,
    }
}

func (dc delimiterChain) Delimiter() rune {
    return dc.delimiter
}

func (dc delimiterChain) Context() interface{} {
    return dc.context
}

type StackItem struct {
    Delimiter rune
    Context Context
}

func (dc delimiterChain) stack(s []StackItem) {
    s[dc.depth] = StackItem{
        Delimiter: dc.delimiter,
        Context: dc.context,
    }

    if dc.parent != nil {
        dc.parent.stack(s)
    }
}

func (dc delimiterChain) Stack() []StackItem {
    s := make([]StackItem, dc.depth + 1)
    dc.stack(s)

    return s
}

func (p *Parser) getContextFromCurrent(dc delimiterChain, dctx map[string]interface{}) Context {
    if dc.Delimiter() == '{' {
        return &ObjectContext{ context: map[string]interface{} { "ObjectKey": dctx["ObjectKey"] } }
    } else {
        return &ListContext{ context: map[string]interface{} { "ListIndex": dctx["ListIndex"] } }
    }
}

// processDelimiter manages the ascending or descending of child structures.
func (p *Parser) processDelimiter(c chan<- interface{}, r rune, dc delimiterChain, dctx map[string]interface{}) (ascend bool, err error) {
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

        context := p.getContextFromCurrent(dc, dctx)
        childDc := dc.Add(r, context)

        err = p.parse(c, childDc)
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

        // fmt.Printf("End of object:\n")
        // for i, si := range dc.Stack() {
        //     indent := strings.Repeat("  ", i + 1)

        //     if i == 0 {
        //         fmt.Printf("%s ROOT\n", indent)
        //     } else {
        //         switch si.Context.(type) {
        //         case *ObjectContext:
        //             fmt.Printf("%s O:%s\n", indent, si.Context.Get("ObjectKey").(string))
        //         case *ListContext:
        //             fmt.Printf("%s L:%d\n", indent, si.Context.Get("ListIndex").(int))
        //         }
        //     }
        // }

        // fmt.Printf("\n")

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

        context := p.getContextFromCurrent(dc, dctx)
        childDc := dc.Add(r, context)

        err = p.parse(c, childDc)
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

func (p *Parser) parse(c chan<- interface{}, dc delimiterChain) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

// TODO(dustin): !! Implement a second channel and a waitgroup for error-handling from the goroutine.

    previousKey := ""

    // Lets us kep track of whether we're on the key or value when processing an
    // object.
    i := 0
    isInObject := dc.Delimiter() == '{'

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

            dctx := map[string]interface{} {
                // Relevant if we're processing through an object.
                "ObjectKey": previousKey,

                // Relevant if we're processing through a list.
                "ListIndex": i,
            }

            ascend, err := p.processDelimiter(c, r, dc, dctx)
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

                    c <- ObjectValue{
                        key: previousKey,
                        value: value,
                    }
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

                    c <- ObjectValue{
                        key: previousKey,
                        value: value,
                    }
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

                        c <- ObjectValue{
                            key: previousKey,
                            value: value,
                        }
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
        }

        i++
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

        dc := delimiterChain{}
        err = p.parse(c, dc)
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
