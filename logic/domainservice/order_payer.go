package domainservice

import (
	"context"
	"errors"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/config"
	"github.com/Ian-zy0329/go-mall/library"
	"github.com/Ian-zy0329/go-mall/logic/do"
)

type OrderPayTemplateContract interface {
	CreateOrderPay() (interface{}, error)
	OrderPayHandlerContract
}

type OrderPayHandlerContract interface {
	CheckRepetition() error
	ValidateOrder() error
	LoadPayAndUserConfig() error
	LoadOrderPayStrategy() error
	HandleOrderPay() (interface{}, error)
}

type OrderPayStrategyContract interface {
	CreatePay(ctx context.Context, order *do.Order, payConfig *OrderPayConfig) (interface{}, error)
}

type OrderPayConfig struct {
	PayUserId   int64
	WxOpenId    string
	WxPayConfig *library.WxPayConfig
	//AliPayConfig
}

type OrderPayTemplate struct {
	OrderPayHandlerContract
}

func (template OrderPayTemplate) CreateOrderPay() (interface{}, error) {
	if err := template.CheckRepetition(); err != nil {
		return nil, err
	}
	if err := template.ValidateOrder(); err != nil {
		return nil, err
	}
	if err := template.LoadPayAndUserConfig(); err != nil {
		return nil, err
	}
	if err := template.LoadOrderPayStrategy(); err != nil {
		return nil, err
	}
	response, err := template.HandleOrderPay()
	if err != nil {
		return nil, err
	}
	return response, nil
}

type CommonOrderPayHandler struct {
	ctx         context.Context
	Scene       string
	UserId      int64
	OrderNo     string
	Order       *do.Order
	PayConfig   *OrderPayConfig
	PayStrategy OrderPayStrategyContract
}

func (handler *CommonOrderPayHandler) CheckRepetition() error {
	return nil
}

func (handler *CommonOrderPayHandler) ValidateOrder() error {
	order, err := NewOrderDomainSvc(handler.ctx).GetSpecifiedUserOrder(handler.OrderNo, handler.UserId)
	if err != nil {
		return err
	}
	if order.OrderStatus > enum.OrderStatusCreated {
		return errcode.ErrOrderParams
	}
	handler.Order = order
	return nil
}

func (handler *CommonOrderPayHandler) LoadPayAndUserConfig() error {
	return nil
}

func (handler *CommonOrderPayHandler) LoadOrderPayStrategy() error {
	return nil
}

func (handler *CommonOrderPayHandler) HandleOrderPay() (interface{}, error) {
	return handler.PayStrategy.CreatePay(handler.ctx, handler.Order, handler.PayConfig)
}

type WxOrderPayHandler struct {
	CommonOrderPayHandler
}

func (wxHandler *WxOrderPayHandler) LoadPayAndUserConfig() error {
	wxHandler.PayConfig.WxPayConfig = &library.WxPayConfig{
		AppId:           config.App.WechatPay.AppId,
		MchId:           config.App.WechatPay.MchId,
		NotifyUrl:       config.App.WechatPay.NotifyUrl,
		PrivateSerialNo: config.App.WechatPay.PrivateSerialNo,
		AesKey:          config.App.WechatPay.AesKey,
	}
	wxHandler.PayConfig.PayUserId = wxHandler.UserId
	openId := "QsudrhgrDYDEEA1344EF"
	wxHandler.PayConfig.WxOpenId = openId
	return nil
}

func (wxHandler *WxOrderPayHandler) LoadOrderPayStrategy() error {
	switch wxHandler.Scene {
	case "app":
	case "jsapi":
		wxHandler.PayStrategy = new(WxJSPayStrategy)
	default:
		return errcode.ErrOrderParams.WithCause(errors.New("unsupported platform"))
	}
	return nil
}

type WxJSPayStrategy struct {
}

func (strategy *WxJSPayStrategy) CreatePay(ctx context.Context, order *do.Order, payConfig *OrderPayConfig) (interface{}, error) {
	wpl := library.NewWxPayLib(ctx, *payConfig.WxPayConfig)
	reply, err := wpl.CreateOrderPay(order, payConfig.WxOpenId)
	if err != nil {
		err = errcode.Wrap("WxJSPayStrategyCreatePayError", err)
	}
	return reply, err
}

func NewOrderPayTemplate(ctx context.Context, userId int64, orderNo, payScene string, payType int) *OrderPayTemplate {
	payTemplate := new(OrderPayTemplate)
	switch payType {
	case enum.PayTypeWxPay:
		payHandler := new(WxOrderPayHandler)
		payHandler.ctx = ctx
		payHandler.UserId = userId
		payHandler.OrderNo = orderNo
		payHandler.Scene = payScene
		payTemplate.OrderPayHandlerContract = payHandler
	}
	return payTemplate
}
