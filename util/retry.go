package util

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"time"
)

const (
	NoProvider      = "No provider available"
	ProviderOffline = "transport: Error while dialing"
	ErrorMessage    = "Rpc provider offline. Retrying %d times ..."
)

type ServiceRetry[Req any, Res any] struct {
}

func (r *ServiceRetry[Req, Res]) CallRpcFunc(ctx context.Context, rpcFunc func(context.Context, Req) (Res, error), rpcReq Req, retryLimit int, retryIntervalMillis int64) (Res, error) {
	var (
		rpcRes Res
		err    error
		retry  = 0
	)
	rpcRes, err = rpcFunc(ctx, rpcReq)
	for err != nil && retry < retryLimit {
		providerError1 := gstr.Pos(err.Error(), NoProvider) >= 0
		providerError2 := gstr.Pos(err.Error(), ProviderOffline) >= 0
		if providerError1 || providerError2 {
			retry++
			g.Log().Errorf(ctx, ErrorMessage, retry)
			// 无限重试
			time.Sleep(time.Duration(retryIntervalMillis) * time.Millisecond)
			rpcRes, err = rpcFunc(ctx, rpcReq)
		} else {
			g.Log().Error(ctx, err)
			retry++
			g.Log().Errorf(ctx, ErrorMessage, retry)
			// 无限重试
			time.Sleep(time.Duration(retryIntervalMillis) * time.Millisecond)
			rpcRes, err = rpcFunc(ctx, rpcReq)
		}
	}
	return rpcRes, err
}
