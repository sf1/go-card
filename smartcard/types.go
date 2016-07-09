package smartcard

import (
    "fmt"
    "bytes"
)

type Reader interface {
    Name() string
    IsCardPresent() bool
    Connect() (Card, error)
}

type Card interface {
    ATR() ATR
    Transmit(command []byte) ([]byte, error)
    Disconnect() error
}

type ATR []byte

func (atr ATR) String() string {
    var buffer bytes.Buffer
    for _, b := range atr {
        buffer.WriteString(fmt.Sprintf("%02x", b))
    }
    return buffer.String()
}
