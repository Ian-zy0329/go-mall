package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/logic/do"
)

func SetDemoOrder(ctx context.Context, demoOrder *do.DemoOrder) error {
	jsonDataBytes, _ := json.Marshal(demoOrder)
	redisKey := fmt.Sprintf(enum.REDIS_KEY_DEMO_ORDER_DETAIL, demoOrder.OrderNo)
	_, err := Redis().Set(ctx, redisKey, jsonDataBytes, 0).Result()
	if err != nil {
		logger.New(ctx).Error("redis set error", "err", err)
		return err
	}
	return nil
}

func GetDemoOrder(ctx context.Context, orderNo string) (*do.DemoOrder, error) {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_DEMO_ORDER_DETAIL, orderNo)
	jsonBytes, err := Redis().Get(ctx, redisKey).Bytes()
	if err != nil {
		logger.New(ctx).Error("redis get error", "err", err)
		return nil, err
	}
	data := new(do.DemoOrder)
	json.Unmarshal(jsonBytes, &data)
	return data, nil
}
