package smartcard

type Reader interface{
    Name() string
    IsCardPresent() bool
    Connect() (Card, error)
}

type Card interface{
    ATR() []byte
    Transmit(command []byte) ([]byte, error)
    Disconnect() error
}

