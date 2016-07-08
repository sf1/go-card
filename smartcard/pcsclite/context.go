package pcsclite

type Context struct {
    client *PCSCDClient
    ctxID uint32
}

func EstablishContext() (*Context, error) {
    var err error
    ctx := &Context{}
    ctx.client, err = PCSCDConnect()
    if err != nil { return nil, err }
    ctx.ctxID, err = ctx.client.EstablishContext()
    return ctx, nil
}

func (ctx *Context) Release() {
    ctx.client.ReleaseContext(ctx.ctxID)
    ctx.client.Close()
}

func (ctx *Context) ListReaders() ([]*Reader, error) {
    readerInfos, err := ctx.client.ListReaders()
    if err != nil { return nil, err }
    result := make([]*Reader, len(readerInfos))
    for i := 0; i < len(readerInfos); i++ {
        result[i] = &Reader{ctx, readerInfos[i]}
    }
    return result, nil
}

func (ctx *Context) WaitReaderChange() error {
    return ctx.client.WaitReaderStateChange()
}
