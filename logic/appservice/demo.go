package appservice

import (
	"context"
	"github.com/Ian-zy0329/go-mall/api/reply"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/cache"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/Ian-zy0329/go-mall/logic/domainservice"
)

// 演示DEMO, 后期使用时删掉

type DemoAppSvc struct {
	ctx           context.Context
	demoDomainSvc *domainservice.DemoDomainSvc
}

func NewDemoAppSvc(ctx context.Context) *DemoAppSvc {
	return &DemoAppSvc{
		ctx:           ctx,
		demoDomainSvc: domainservice.NewDemoDomainSvc(ctx),
	}
}

//func (das *DemoAppSvc)DoSomething() {
//	demo, err := das.demoDomainSvc.GetDemoEntity(id)
//	if err != nil {
//		logger.New(das.ctx).Error("DemoAppSvc DoSomething err", err)
//		return err
//	}
//	......
//}

// GetDemoIdentities 配置GORM时的演示方法, 显的有点脑残,
// 后面章节再解释怎么用ApplicationService 进行逻辑解耦
func (das *DemoAppSvc) GetDemoIdentities() ([]int64, error) {
	demos, err := das.demoDomainSvc.GetDemos()
	if err != nil {
		return nil, err
	}
	identities := make([]int64, 0, len(demos))

	for _, demo := range demos {
		identities = append(identities, demo.Id)
	}
	return identities, nil
}

func (das *DemoAppSvc) CreateDemoOrder(orderRequest *request.DemoOrderCreate) (*reply.DemoOrder, error) {
	demoOrder := new(do.DemoOrder)
	err := util.CopyProperties(demoOrder, orderRequest)
	if err != nil {
		errcode.Wrap("请求转换成demoOrderDo失败", err)
		return nil, err
	}
	demoOrderDo, err := das.demoDomainSvc.CreateDemoOrder(demoOrder)
	if err != nil {
		err = errcode.Wrap("创建demoOrderDo失败", err)
		return nil, err
	}

	cache.SetDemoOrder(das.ctx, demoOrderDo)
	cacheData, _ := cache.GetDemoOrder(das.ctx, demoOrderDo.OrderNo)
	logger.New(das.ctx).Info("cacheData", cacheData)
	replyDemoOrder := new(reply.DemoOrder)
	err = util.CopyProperties(replyDemoOrder, demoOrderDo)
	if err != nil {
		errcode.Wrap("demoOrderDo转换成replyDemoOrder失败", err)
		return nil, err
	}
	return replyDemoOrder, nil
}

func (das *DemoAppSvc) InitCommodityCategoryData() error {
	cds := domainservice.NewCommodityDomainSvc(das.ctx)
	err := cds.InitCategoryData()
	return err
}

func (das *DemoAppSvc) InitCommodityData() error {
	cds := domainservice.NewCommodityDomainSvc(das.ctx)
	err := cds.InitCommodityData()
	return err
}
