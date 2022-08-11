package rpc

import (
	"github.com/WesleyWu/gf-dubbogo/util/dubbogo"
	"github.com/gogf/gf/v2/os/gctx"
)

func init() {
	ctx := gctx.New()
	err := dubbogo.StartConsumersByCfg(ctx, false, 180)
	if err != nil {
		panic(err)
	}
}
