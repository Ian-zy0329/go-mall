package library

import (
	"context"
	"encoding/json"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util/httptool"
)

type WhoisLib struct {
	ctx context.Context
}

func NewWhoisLib(ctx context.Context) *WhoisLib {
	return &WhoisLib{
		ctx: ctx,
	}
}

type WhoisDetail struct {
	Ip      string `json:"ip"`
	Success string `json:"success"`
}

func (whois *WhoisLib) GetHostIpDetail() (*WhoisDetail, error) {
	log := logger.New(whois.ctx)
	httpStatusCode, respBody, err := httptool.Get(whois.ctx, "https://ipwho.is", httptool.WithHeaders(map[string]string{"User-Agent": "curl/7.77.0"}))
	if err != nil {
		log.Error("whois request error", "err", err, "httpStatusCode", httpStatusCode, "respBody", string(respBody))
		return nil, err
	}
	reply := new(WhoisDetail)
	json.Unmarshal(respBody, reply)
	return reply, nil
}
