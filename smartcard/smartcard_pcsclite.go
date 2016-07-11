// +build !windows

package smartcard

import "github.com/sf1/go-card/smartcard/pcsc"

type Context struct {
    client *pcsc.PCSCLiteClient
    ctxID uint32
}

func EstablishContext() (*Context, error) {
    var err error
    context := &Context{}
    context.client, err = pcsc.PCSCLiteConnect()
    if err != nil { return nil, err }
    context.ctxID, err = context.client.EstablishContext()
    return context, nil
}

func (ctx *Context) Release() error {
    return ctx.client.ReleaseContext(ctx.ctxID)
}

func (ctx *Context) ListReaders() ([]*Reader, error) {
    return ctx.listReaders(false)
}

func (ctx *Context) ListReadersWithCard() ([]*Reader, error) {
    return ctx.listReaders(true)
}

func (ctx *Context) listReaders(withCard bool) ([]*Reader, error) {
    readerInfos, err := ctx.client.ListReaders()
    if err != nil { return nil, err }
    result := make([]*Reader, 0, len(readerInfos))
    for i := 0; i < len(readerInfos); i++ {
        if withCard {
            if readerInfos[i].IsCardPresent() {
                result = append(result, &Reader{ctx, *readerInfos[i]})
            }
        } else {
            result = append(result, &Reader{ctx, *readerInfos[i]})
        }
    }
    return result, nil
}

func (ctx *Context) WaitForCardPresent() (*Reader, error) {
    var reader *Reader
    for reader == nil {
        count, err := ctx.client.SyncReaderStates()
        if err != nil { return nil, err}
        for i := uint32(0); i < count; i++ {
            state := ctx.client.ReaderStates()[i]
            if state.IsCardPresent() {
                reader = &Reader{ctx, state}
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

type Reader struct {
    context *Context
    info pcsc.ReaderInfo
}

func (r *Reader) Name() string {
    return r.info.Name()
}

func (r *Reader) IsCardPresent() bool {
    r.context.client.SyncReaderStates()
    return r.info.IsCardPresent()
}

func (r *Reader) Connect() (*Card, error) {
    cardID, protocol, err := r.context.client.CardConnect(
        r.context.ctxID, r.info.Name())
    if err != nil { return nil, err }
    return &Card{
        r.context,
        cardID,
        protocol,
        r.info.CardAtr[:r.info.CardAtrLength],
    }, nil
}

type Card struct {
    context *Context
    cardID int32
    protocol uint32
    atr ATR
}

func (c *Card) ATR() ATR {
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
    err := c.context.client.CardDisconnect(c.cardID)
    if err != nil { return err }
    return nil
}
