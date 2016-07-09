package pcsclite

type Reader struct {
    info ReaderInfo
}

func (r *Reader) Name() string {
    return r.info.Name()
}

func (r *Reader) IsCardPresent() bool {
    return r.info.IsCardPresent()
}

func (r *Reader) Connect() (*Card, error) {
    client, ctx, err := r.establishContext()
    if err != nil { return nil, err }
    cardID, protocol, err := client.CardConnect(ctx, r.info.Name())
    if err != nil { return nil, err }
    return &Card{
        client,
        ctx,
        cardID,
        protocol,
        r.info.CardAtr[:r.info.CardAtrLength],
    }, nil
}

func (r *Reader) establishContext() (*PCSCDClient, uint32, error) {
    client, err := PCSCDConnect()
    if err != nil { return nil, 0, err }
    ctx, err := client.EstablishContext()
    return client, ctx, nil
}
