package controller

import (
	"errors"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/logic/appservice"
	"github.com/gin-gonic/gin"
)

func OrderCreate(c *gin.Context) {
	request := new(request.OrderCreate)
	if err := c.ShouldBindJSON(request); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	orderSvc := appservice.NewOrderAppSvc(c)
	reply, err := orderSvc.CreateOrder(request, c.GetInt64("userId"))
	if err != nil {
		if errors.Is(err, errcode.ErrCartItemParam) {
			app.NewResponse(c).Error(errcode.ErrCartItemParam)
		} else if errors.Is(err, errcode.ErrCartWrongUser) {
			app.NewResponse(c).Error(errcode.ErrCartWrongUser)
		} else if errors.Is(err, errcode.ErrCommodityStockOut) {
			app.NewResponse(c).Error(errcode.ErrCommodityStockOut.WithCause(err))
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}
	app.NewResponse(c).Success(reply)
}

func UserOrders(c *gin.Context) {
	pagination := app.NewPagination(c)
	orderAppSvc := appservice.NewOrderAppSvc(c)
	replyOrders, err := orderAppSvc.GetUserOrders(c.GetInt64("userId"), pagination)
	if err != nil {
		app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
	}
	app.NewResponse(c).SetPagination(pagination).Success(replyOrders)
}

func OrderInfo(c *gin.Context) {
	orderNo := c.Param("order_no")
	orderAppSvc := appservice.NewOrderAppSvc(c)
	replyOrder, err := orderAppSvc.GetOrderInfo(orderNo, c.GetInt64("userId"))
	if err != nil {
		if errors.Is(err, errcode.ErrOrderParams) {
			app.NewResponse(c).Error(errcode.ErrOrderParams)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}

	app.NewResponse(c).Success(replyOrder)
}

func CancelOrder(c *gin.Context) {
	orderNo := c.Param("order_no")
	orderAppSvc := appservice.NewOrderAppSvc(c)
	err := orderAppSvc.CancelOrder(orderNo, c.GetInt64("userId"))
	if err != nil {
		if errors.Is(err, errcode.ErrOrderParams) {
			app.NewResponse(c).Error(errcode.ErrOrderParams)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}
	app.NewResponse(c).SuccessOk()
}
