package pcsclite

type Card struct {
    client *PCSCDClient
    context uint32
    cardID int32
    protocol uint32
    atr []byte
}

func (c *Card) ATR() []byte {
    return c.atr
}

func (c *Card) Transmit(command []byte) ([]byte, error) {
    response := make([]byte, 258)
    received, err := c.client.Transmit(c.cardID, c.protocol,
        command, response)
    if err != nil { return nil, err }
    return response[:received], nil
}

func (c *Card) Disconnect() error {
    err := c.client.CardDisconnect(c.cardID)
    if err != nil { return err }
    err = c.client.ReleaseContext(c.context)
    if err != nil { return err }
    c.client.Close()
    return nil
}

