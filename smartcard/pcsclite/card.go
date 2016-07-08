package pcsclite

type Card struct {
    context *Context
    cardID int32
    protocol uint32
    atr []byte
}

func (c *Card) ATR() []byte {
    return c.atr
}

func (c *Card) Transmit(command []byte) ([]byte, error) {
    response := make([]byte, 258)
    received, err := c.context.client.Transmit(c.cardID, c.protocol,
        command, response)
    if err != nil { return nil, err }
    return response[:received], nil
}

func (c *Card) Disconnect() error {
    return c.context.client.CardDisconnect(c.cardID)
}
