package pcsclite

type Reader struct {
    context *Context
    info *ReaderInfo
}

func (r *Reader) Name() string {
    return r.info.Name()
}

func (r *Reader) IsCardPresent() bool {
    //r.context.client.SyncReaderStates()
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
