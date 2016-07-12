// +build !windows

package smartcard

import "github.com/sf1/go-card/smartcard/pcsc"

// A smart card context is required to access readers and cards.
type Context struct {
    client *pcsc.PCSCLiteClient
    ctxID uint32
}

// Establish smart card context.
// This should be the first function to be called.
func EstablishContext() (*Context, error) {
    var err error
    context := &Context{}
    context.client, err = pcsc.PCSCLiteConnect()
    if err != nil { return nil, err }
    context.ctxID, err = context.client.EstablishContext()
    return context, nil
}

// Release resources associated with smart card context.
func (ctx *Context) Release() error {
    return ctx.client.ReleaseContext(ctx.ctxID)
}

// List all smart card readers.
func (ctx *Context) ListReaders() ([]*Reader, error) {
    return ctx.listReaders(false)
}

// List smart card readers with inserted cards.
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

// Block until a smart card is inserted into any reader.
// Returns immediately if card already present.
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

// Smart card reader. 
// Note that physical card readers with slots for multiple cards are
// represented by one Reader instance per slot.
type Reader struct {
    context *Context
    info pcsc.ReaderInfo
}

// Return name of card reader.
func (r *Reader) Name() string {
    return r.info.Name()
}

// Check if card is present.
func (r *Reader) IsCardPresent() bool {
    r.context.client.SyncReaderStates()
    return r.info.IsCardPresent()
}

// Connect to card.
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

// Smart card.
type Card struct {
    context *Context
    cardID int32
    protocol uint32
    atr ATR
}

// Return card ATR (answer to reset).
func (c *Card) ATR() ATR {
    return c.atr
}

// Trasmit bytes to card and return response.
func (c *Card) Transmit(command []byte) ([]byte, error) {
    response := make([]byte, 258)
    received, err := c.context.client.Transmit(c.cardID, c.protocol,
        command, response)
    if err != nil { return nil, err }
    return response[:received], nil
}

// Disconnect from card.
func (c *Card) Disconnect() error {
    err := c.context.client.CardDisconnect(c.cardID)
    if err != nil { return err }
    return nil
}
