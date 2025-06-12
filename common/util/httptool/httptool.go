package httptool

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"io/ioutil"
	"net/http"
	"time"
)

type requestOption struct {
	ctx     context.Context
	timeout time.Duration
	data    []byte
	headers map[string]string
}

func defaultRequestOptions() *requestOption {
	return &requestOption{
		ctx:     context.Background(),
		timeout: 5 * time.Second,
		data:    nil,
		headers: map[string]string{},
	}
}

type Option interface {
	apply(option *requestOption) error
}

type optionFunc func(option *requestOption) error

func (f optionFunc) apply(option *requestOption) error {
	return f(option)
}

func WithContext(ctx context.Context) Option {
	return optionFunc(func(option *requestOption) (err error) {
		option.ctx = ctx
		return
	})
}

func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(option *requestOption) (err error) {
		option.timeout, err = timeout, nil
		return
	})
}

func WithHeaders(headers map[string]string) Option {
	return optionFunc(func(option *requestOption) (err error) {
		for k, v := range headers {
			option.headers[k] = v
		}
		return
	})
}

func WithData(data []byte) Option {
	return optionFunc(func(option *requestOption) (err error) {
		option.data, err = data, nil
		return
	})
}

func Get(ctx context.Context, url string, opts ...Option) (httpStatusCode int, respBody []byte, err error) {
	opts = append(opts, WithContext(ctx))
	return Request("GET", url, opts...)
}

func Post(ctx context.Context, url string, data []byte, opts ...Option) (httpStatusCode int, respBody []byte, err error) {
	defaultHeader := map[string]string{"Content-Type": "application/json"}
	var newOptions []Option
	newOptions = append(newOptions, WithHeaders(defaultHeader), WithData(data), WithContext(ctx))
	newOptions = append(newOptions, opts...)
	httpStatusCode, respBody, err = Request("POST", url, newOptions...)
	return
}

func Request(method string, url string, options ...Option) (httpStatusCode int, respBody []byte, err error) {
	start := time.Now()
	reqOpts := defaultRequestOptions()
	for _, opt := range options {
		err = opt.apply(reqOpts)
		if err != nil {
			return
		}
	}
	log := logger.New(reqOpts.ctx)
	defer func() {
		if err != nil {
			log.Error("HTTP_REQUEST_ERROR_LOG", "method", method, "url", url, "body", reqOpts.data, "reply", respBody, "err", err)
		}
	}()

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqOpts.data))
	if err != nil {
		return
	}
	req = req.WithContext(reqOpts.ctx)
	defer req.Body.Close()

	traceId, spanId, _ := util.GetTraceInfoFromCtx(reqOpts.ctx)
	reqOpts.headers["traceid"] = traceId
	reqOpts.headers["spanid"] = spanId
	if len(reqOpts.headers) != 0 {
		for k, v := range reqOpts.headers {
			req.Header.Add(k, v)
		}
	}

	client := &http.Client{Timeout: reqOpts.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	dur := time.Since(start).Milliseconds()
	if dur >= 3000 {
		log.Warn("HTTP_REQUEST_SLOW_LOG", "method", method, "url", url, "body", reqOpts.data, "reply", respBody, "dur/ms", dur)
	} else {
		log.Debug("HTTP_REQUEST_LOG", "method", method, "url", url, "body", reqOpts.data, "reply", respBody, "dur/ms", dur)
	}

	httpStatusCode = resp.StatusCode
	if httpStatusCode != http.StatusOK {
		err = errcode.Wrap("request api error", errors.New(fmt.Sprintf("non 200 response, response code: %d", httpStatusCode)))
		return
	}

	respBody, _ = ioutil.ReadAll(resp.Body)
	return
}
