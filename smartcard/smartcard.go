package smartcard

import (
    "fmt"
    "bytes"
)

type ATR []byte

func (atr ATR) String() string {
    var buffer bytes.Buffer
    for _, b := range atr {
        buffer.WriteString(fmt.Sprintf("%02x", b))
    }
    return buffer.String()
}
