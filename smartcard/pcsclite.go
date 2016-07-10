// +build !windows

package smartcard

type PCSCLiteContext struct {
    client *PCSCLiteClient
    ctxID uint32
}

func EstablishContext() (Context, error) {
    var err error
    context := &PCSCLiteContext{}
    context.client, err = PCSCLiteConnect()
    if err != nil { return nil, err }
    context.ctxID, err = context.client.EstablishContext()
    return context, nil
}

func (ctx *PCSCLiteContext) Release() error {
    return ctx.client.ReleaseContext(ctx.ctxID)
}

func (ctx *PCSCLiteContext) ListReaders() ([]Reader, error) {
    return ctx.listReaders(false)
}

func (ctx *PCSCLiteContext) ListReadersWithCard() ([]Reader, error) {
    return ctx.listReaders(true)
}

func (ctx *PCSCLiteContext) listReaders(withCard bool) ([]Reader, error) {
    readerInfos, err := ctx.client.ListReaders()
    if err != nil { return nil, err }
    result := make([]Reader, 0, len(readerInfos))
    for i := 0; i < len(readerInfos); i++ {
        if withCard {
            if readerInfos[i].IsCardPresent() {
                result = append(result, &PCSCLiteReader{ctx, *readerInfos[i]})
            }
        } else {
            result = append(result, &PCSCLiteReader{ctx, *readerInfos[i]})
        }
    }
    return result, nil
}

func (ctx *PCSCLiteContext) WaitForCardPresent() (Reader, error) {
    var reader *PCSCLiteReader
    for reader == nil {
        count, err := ctx.client.SyncReaderStates()
        if err != nil { return nil, err}
        for i := uint32(0); i < count; i++ {
            if ctx.client.readerStates[i].IsCardPresent() {
                reader = &PCSCLiteReader{ctx, ctx.client.readerStates[i]}
                break
            }
        }
        if reader != nil {
            break
        }
        ctx.client.WaitReaderStateChange()
    }
    return reader, nil
}

type PCSCLiteReader struct {
    context *PCSCLiteContext
    info ReaderInfo
}

func (r *PCSCLiteReader) Name() string {
    return r.info.Name()
}

func (r *PCSCLiteReader) IsCardPresent() bool {
    return r.info.IsCardPresent()
}

func (r *PCSCLiteReader) Connect() (Card, error) {
    cardID, protocol, err := r.context.client.CardConnect(
        r.context.ctxID, r.info.Name())
    if err != nil { return nil, err }
    return &PCSCLiteCard{
        r.context,
        cardID,
        protocol,
        r.info.CardAtr[:r.info.CardAtrLength],
    }, nil
}

type PCSCLiteCard struct {
    context *PCSCLiteContext
    cardID int32
    protocol uint32
    atr ATR
}

func (c *PCSCLiteCard) ATR() ATR {
    return c.atr
}

func (c *PCSCLiteCard) Transmit(command []byte) ([]byte, error) {
    response := make([]byte, 258)
    received, err := c.context.client.Transmit(c.cardID, c.protocol,
        command, response)
    if err != nil { return nil, err }
    return response[:received], nil
}

func (c *PCSCLiteCard) Disconnect() error {
    err := c.context.client.CardDisconnect(c.cardID)
    if err != nil { return err }
    return nil
}
