package controller

import (
	"errors"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/logic/appservice"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"strconv"
)

func AddCartItem(c *gin.Context) {
	request := new(request.AddCartItem)
	if err := c.ShouldBindJSON(request); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}

	svc := appservice.NewCartAppSvc(c)
	err := svc.AddCartItem(request, c.GetInt64("userId"))
	if err != nil {
		if errors.Is(err, errcode.ErrCommodityNotExists) {
			app.NewResponse(c).Error(errcode.ErrCommodityNotExists)
		} else if errors.Is(err, errcode.ErrCommodityStockOut) {
			app.NewResponse(c).Error(errcode.ErrCommodityStockOut)
		} else {
			// WithCause 记得加, 不然请求的错误日志里记不到错误原因
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}
	app.NewResponse(c).SuccessOk()
}

func CheckCartItemBill(c *gin.Context) {
	itemIdList := c.QueryArray("item_id")
	if len(itemIdList) == 0 {
		app.NewResponse(c).Error(errcode.ErrParams)
	}
	itemIds := lo.Map(itemIdList, func(itemId string, index int) int64 {
		i, _ := strconv.ParseInt(itemId, 10, 64)
		return i
	})
	cartAppSvc := appservice.NewCartAppSvc(c)
	replyData, err := cartAppSvc.CheckCartItemBillV2(itemIds, c.GetInt64("userId"))
	if err != nil {
		if errors.Is(err, errcode.ErrCartItemParam) {
			app.NewResponse(c).Error(errcode.ErrCartItemParam)
		} else if errors.Is(err, errcode.ErrCartWrongUser) {
			app.NewResponse(c).Error(errcode.ErrCartWrongUser)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}
	app.NewResponse(c).Success(replyData)
}

func UpdateCartItem(c *gin.Context) {
	request := new(request.CartItemUpdate)
	if err := c.ShouldBindJSON(request); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	cartAppSvc := appservice.NewCartAppSvc(c)
	err := cartAppSvc.UpdateCartItem(request, c.GetInt64("userId"))
	if err != nil {
		if errors.Is(err, errcode.ErrCartItemParam) {
			app.NewResponse(c).Error(errcode.ErrCartItemParam)
		} else if errors.Is(err, errcode.ErrCartWrongUser) {
			app.NewResponse(c).Error(errcode.ErrCartWrongUser)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}
	app.NewResponse(c).SuccessOk()
}

func UserCartItems(c *gin.Context) {
	cartAppSvc := appservice.NewCartAppSvc(c)
	cartItems, err := cartAppSvc.GetUserCartItems(c.GetInt64("userId"))
	if err != nil {
		app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		return
	}
	app.NewResponse(c).Success(cartItems)
}

func DeleteCartItem(c *gin.Context) {
	itemId, _ := strconv.ParseInt(c.Param("item_id"), 10, 64)
	cartAppSvc := appservice.NewCartAppSvc(c)
	err := cartAppSvc.DeleteCartItem(c.GetInt64("userId"), itemId)
	if err != nil {
		if errors.Is(err, errcode.ErrParams) {
			app.NewResponse(c).Error(errcode.ErrParams)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}
	app.NewResponse(c).SuccessOk()
}
