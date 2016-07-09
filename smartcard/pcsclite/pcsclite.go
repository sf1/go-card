package pcsclite

type PCSCLiteReader struct {
    info ReaderInfo
}

func (r *PCSCLiteReader) Name() string {
    return r.info.Name()
}

func (r *PCSCLiteReader) IsCardPresent() bool {
    return r.info.IsCardPresent()
}

func (r *PCSCLiteReader) Connect() (Card, error) {
    client, ctx, err := r.establishContext()
    if err != nil { return nil, err }
    cardID, protocol, err := client.CardConnect(ctx, r.info.Name())
    if err != nil { return nil, err }
    return &PCSCLiteCard{
        client,
        ctx,
        cardID,
        protocol,
        r.info.CardAtr[:r.info.CardAtrLength],
    }, nil
}

func (r *PCSCLiteReader) establishContext() (*PCSCLiteClient, uint32, error) {
    client, err := PCSCLiteConnect()
    if err != nil { return nil, 0, err }
    ctx, err := client.EstablishContext()
    return client, ctx, nil
}

type PCSCLiteCard struct {
    client *PCSCLiteClient
    context uint32
    cardID int32
    protocol uint32
    atr []byte
}

func (c *PCSCLiteCard) ATR() []byte {
    return c.atr
}

func (c *PCSCLiteCard) Transmit(command []byte) ([]byte, error) {
    response := make([]byte, 258)
    received, err := c.client.Transmit(c.cardID, c.protocol,
        command, response)
    if err != nil { return nil, err }
    return response[:received], nil
}

func (c *PCSCLiteCard) Disconnect() error {
    err := c.client.CardDisconnect(c.cardID)
    if err != nil { return err }
    err = c.client.ReleaseContext(c.context)
    if err != nil { return err }
    c.client.Close()
    return nil
}

func ListReaders() ([]Reader, error) {
    return listReaders(false)
}

func ListReadersWithCard() ([]Reader, error) {
    return listReaders(true)
}

func listReaders(withCard bool) ([]Reader, error) {
    client, err := PCSCLiteConnect()
    if err != nil { return nil, err}
    defer client.Close()
    readerInfos, err := client.ListReaders()
    if err != nil { return nil, err }
    result := make([]Reader, 0, len(readerInfos))
    for i := 0; i < len(readerInfos); i++ {
        if withCard {
            if readerInfos[i].IsCardPresent() {
                result = append(result, &PCSCLiteReader{*readerInfos[i]})
            }
        } else {
            result = append(result, &PCSCLiteReader{*readerInfos[i]})
        }
    }
    return result, nil
}

func WaitForCardPresent() (Reader, error) {
    var reader *PCSCLiteReader
    client, err := PCSCLiteConnect()
    if err != nil { return nil, err}
    defer client.Close()
    for reader == nil {
        count, err := client.SyncReaderStates()
        if err != nil { return nil, err}
        for i := uint32(0); i < count; i++ {
            if client.readerStates[i].IsCardPresent() {
                reader = &PCSCLiteReader{client.readerStates[i]}
                break
            }
        }
        if reader != nil {
            break
        }
        client.WaitReaderStateChange()
    }
    return reader, nil
}
