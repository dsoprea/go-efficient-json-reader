package jsonreader

import (
    "testing"
    "os"
    "path"
    "fmt"
    "io"
    "strings"
    "sort"

    "github.com/dsoprea/go-logging"
)

var (
    testingAssetsPath = ""
)

func flattenStream(r io.Reader) (ts []string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(err)
        }
    }()

    c := make(chan interface{}, 0)

    p := NewParser(r)
    err = p.Parse(c)
    log.PanicIf(err)

    ts = make([]string, 0)
    for token := range c {
        flat := ""

        switch token.(type) {
        case ObjectOpen:
            flat = "/OBJECTOPEN"
        case ObjectClose:
            flat = "/OBJECTCLOSE"
        case ListOpen:
            flat = "/LISTOPEN"
        case ListClose:
            flat = "/LISTCLOSE"
        case ObjectKey:
            flat = fmt.Sprintf(":%s", token)
        case ObjectValue:
            ov := token.(ObjectValue)
            v := ov.Value()

            switch v.(type) {
            case float64:
                flat = fmt.Sprintf("[%s] F %f", ov.Key(), v)
            case int64:
                flat = fmt.Sprintf("[%s] I %d", ov.Key(), v)
            case string:
                flat = fmt.Sprintf("[%s] S %s", ov.Key(), v)
            }
        case float64:
            flat = fmt.Sprintf("#FLOAT64=%f", token)
        case int64:
            flat = fmt.Sprintf("#INT64=%d", token)
        case string:
            flat = fmt.Sprintf("#STRING=%s", token)
        case SimpleObject:
            // Produce adeterministic string representation by sorting the
            // keys.

            o := map[string]interface{}(token.(SimpleObject))

            keys := make([]string, len(o))
            i := 0
            for k, _ := range o {
                keys[i] = k
                i++
            }

            ss := sort.StringSlice(keys)
            ss.Sort()

            couplets := make([]string, len(o))
            for j, k := range ss {
                couplets[j] = fmt.Sprintf("%s:%v", k, o[k])
            }

            flat = fmt.Sprintf("@%s", strings.Join(couplets, " "))
        }

        ts = append(ts, flat)
    }

    return ts, nil
}

func TestParseData1(t *testing.T) {
    filepath := path.Join(testingAssetsPath, "data1.json")
    f, err := os.Open(filepath)

    defer f.Close()

    ts, err := flattenStream(f)
    log.PanicIf(err)

    expected := []string{
        "/OBJECTOPEN",
        ":locations",
        "/LISTOPEN",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1517218739237",
        ":latitudeE7",
        "[latitudeE7] F 265620925.000000",
        ":longitudeE7",
        "[longitudeE7] F -801004559.000000",
        ":accuracy",
        "[accuracy] F 600.000000",
        "/OBJECTCLOSE",
        "@accuracy:600 latitudeE7:2.65620925e+08 longitudeE7:-8.01004559e+08 timestampMs:1517218739237",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1517218624576",
        ":latitudeE7",
        "[latitudeE7] F 265620925.000000",
        ":longitudeE7",
        "[longitudeE7] F -801004559.000000",
        ":accuracy",
        "[accuracy] F 600.000000",
        "/OBJECTCLOSE",
        "@accuracy:600 latitudeE7:2.65620925e+08 longitudeE7:-8.01004559e+08 timestampMs:1517218624576",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1517218591314",
        ":latitudeE7",
        "[latitudeE7] F 265692905.000000",
        ":longitudeE7",
        "[longitudeE7] F -800990596.000000",
        ":accuracy",
        "[accuracy] F 1500.000000",
        "/OBJECTCLOSE",
        "@accuracy:1500 latitudeE7:2.65692905e+08 longitudeE7:-8.00990596e+08 timestampMs:1517218591314",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1517218469293",
        ":latitudeE7",
        "[latitudeE7] F 265625602.000000",
        ":longitudeE7",
        "[longitudeE7] F -801018779.000000",
        ":accuracy",
        "[accuracy] F 8.000000",
        ":velocity",
        "[velocity] F 0.000000",
        ":altitude",
        "[altitude] F -10.000000",
        ":verticalAccuracy",
        "[verticalAccuracy] F 16.000000",
        "/OBJECTCLOSE",
        "@accuracy:8 altitude:-10 latitudeE7:2.65625602e+08 longitudeE7:-8.01018779e+08 timestampMs:1517218469293 velocity:0 verticalAccuracy:16",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1517218439062",
        ":latitudeE7",
        "[latitudeE7] F 265620925.000000",
        ":longitudeE7",
        "[longitudeE7] F -801004559.000000",
        ":accuracy",
        "[accuracy] F 600.000000",
        "/OBJECTCLOSE",
        "@accuracy:600 latitudeE7:2.65620925e+08 longitudeE7:-8.01004559e+08 timestampMs:1517218439062",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1517218384598",
        ":latitudeE7",
        "[latitudeE7] F 265620925.000000",
        ":longitudeE7",
        "[longitudeE7] F -801004559.000000",
        ":accuracy",
        "[accuracy] F 600.000000",
        "/OBJECTCLOSE",
        "@accuracy:600 latitudeE7:2.65620925e+08 longitudeE7:-8.01004559e+08 timestampMs:1517218384598",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1517218204331",
        ":latitudeE7",
        "[latitudeE7] F 265625020.000000",
        ":longitudeE7",
        "[longitudeE7] F -801020095.000000",
        ":accuracy",
        "[accuracy] F 16.000000",
        ":velocity",
        "[velocity] F 0.000000",
        ":heading",
        "[heading] F 216.000000",
        ":altitude",
        "[altitude] F 21.000000",
        ":verticalAccuracy",
        "[verticalAccuracy] F 32.000000",
        "/OBJECTCLOSE",
        "@accuracy:16 altitude:21 heading:216 latitudeE7:2.6562502e+08 longitudeE7:-8.01020095e+08 timestampMs:1517218204331 velocity:0 verticalAccuracy:32",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1376344387457",
        ":latitudeE7",
        "[latitudeE7] F 265515950.000000",
        ":longitudeE7",
        "[longitudeE7] F -800931448.000000",
        ":accuracy",
        "[accuracy] F 9.000000",
        ":activity",
        "/LISTOPEN",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1376344387457",
        ":activity",
        "/LISTOPEN",
        "/OBJECTOPEN",
        ":type",
        "[type] S UNKNOWN",
        ":confidence",
        "[confidence] F 60.000000",
        "/OBJECTCLOSE",
        "@confidence:60 type:UNKNOWN",
        "/OBJECTOPEN",
        ":type",
        "[type] S STILL",
        ":confidence",
        "[confidence] F 32.000000",
        "/OBJECTCLOSE",
        "@confidence:32 type:STILL",
        "/OBJECTOPEN",
        ":type",
        "[type] S IN_VEHICLE",
        ":confidence",
        "[confidence] F 7.000000",
        "/OBJECTCLOSE",
        "@confidence:7 type:IN_VEHICLE",
        "/LISTCLOSE",
        "/OBJECTCLOSE",
        "@timestampMs:1376344387457",
        "/LISTCLOSE",
        "/OBJECTCLOSE",
        "@accuracy:9 latitudeE7:2.6551595e+08 longitudeE7:-8.00931448e+08 timestampMs:1376344387457",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1376344381236",
        ":latitudeE7",
        "[latitudeE7] F 265515931.000000",
        ":longitudeE7",
        "[longitudeE7] F -800931460.000000",
        ":accuracy",
        "[accuracy] F 10.000000",
        "/OBJECTCLOSE",
        "@accuracy:10 latitudeE7:2.65515931e+08 longitudeE7:-8.0093146e+08 timestampMs:1376344381236",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1376344376523",
        ":latitudeE7",
        "[latitudeE7] F 265515909.000000",
        ":longitudeE7",
        "[longitudeE7] F -800931443.000000",
        ":accuracy",
        "[accuracy] F 11.000000",
        "/OBJECTCLOSE",
        "@accuracy:11 latitudeE7:2.65515909e+08 longitudeE7:-8.00931443e+08 timestampMs:1376344376523",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1376344369732",
        ":latitudeE7",
        "[latitudeE7] F 265515908.000000",
        ":longitudeE7",
        "[longitudeE7] F -800931446.000000",
        ":accuracy",
        "[accuracy] F 14.000000",
        "/OBJECTCLOSE",
        "@accuracy:14 latitudeE7:2.65515908e+08 longitudeE7:-8.00931446e+08 timestampMs:1376344369732",
        "/OBJECTOPEN",
        ":timestampMs",
        "[timestampMs] S 1376344184115",
        ":latitudeE7",
        "[latitudeE7] F 265515889.000000",
        ":longitudeE7",
        "[longitudeE7] F -800931423.000000",
        ":accuracy",
        "[accuracy] F 20.000000",
        "/OBJECTCLOSE",
        "@accuracy:20 latitudeE7:2.65515889e+08 longitudeE7:-8.00931423e+08 timestampMs:1376344184115",
        "/LISTCLOSE",
        "/OBJECTCLOSE",
        "@",
    }

    for i, entry := range ts {
        if expected[i] != entry {
            t.Fatalf("Item (%d) value [%s] should be [%s].", i, entry, expected[i])
        }
    }
}

func TestParseData2(t *testing.T) {
    filepath := path.Join(testingAssetsPath, "data2.json")
    f, err := os.Open(filepath)

    defer f.Close()

    ts, err := flattenStream(f)
    log.PanicIf(err)

    expected := []string{
        "/LISTOPEN",
        "#FLOAT64=99.000000",
        "#FLOAT64=123.450000",
        "/LISTOPEN",
        "#FLOAT64=678.900000",
        "/LISTCLOSE",
        "#STRING=test string",
        "/LISTCLOSE",
    }

    for i, entry := range ts {
        if expected[i] != entry {
            t.Fatalf("Item (%d) value [%s] should be [%s].", i, entry, expected[i])
        }
    }
}

func TestParseScalar1(t *testing.T) {
    filepath := path.Join(testingAssetsPath, "scalar1.json")
    f, err := os.Open(filepath)

    defer f.Close()

    ts, err := flattenStream(f)
    log.PanicIf(err)

    expected := []string{
        "#FLOAT64=1234.567800",
    }

    for i, entry := range ts {
        if expected[i] != entry {
            t.Fatalf("Item (%d) value [%s] should be [%s].", i, entry, expected[i])
        }
    }
}

func init() {
    goPath := os.Getenv("GOPATH")
    testingAssetsPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-efficient-json-reader", "testing")
}
