package dao

import (
	"context"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/model"
	"github.com/Ian-zy0329/go-mall/logic/do"
)

type DemoDao struct {
	ctx context.Context
}

func NewDemoDao(ctx context.Context) *DemoDao {
	return &DemoDao{ctx: ctx}
}

func (d *DemoDao) GetAllDemos() (demos []*model.DemoOrder, err error) {
	err = DB().WithContext(d.ctx).Find(&demos).Error
	if err != nil {
		return nil, err
	}
	return demos, nil
}

func (demo *DemoDao) CreateDemoOrder(demoOrder *do.DemoOrder) (*model.DemoOrder, error) {
	model := new(model.DemoOrder)
	err := util.CopyProperties(model, demoOrder)
	if err != nil {
		return nil, err
	}
	err = DB().WithContext(demo.ctx).Create(model).Error
	return model, err
}
