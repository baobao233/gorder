package discovery

import (
	"context"
	"github.com/baobao233/gorder/common/discovery/consul"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

func RegisterToConsul(ctx context.Context, serviceName string) (func() error, error) {
	registry, err := consul.New(viper.GetString("consul-addr"))
	if err != nil {
		return func() error {
			return nil
		}, err
	}
	instanceID := GenerateInstanceID(serviceName)
	grpcAddr := viper.Sub(serviceName).GetString("grpc-addr")
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddr); err != nil {
		return func() error {
			return nil
		}, err
	}
	// 成功注册的时候开启协程进行 check heartbeat
	go func() {
		// 一定要有个死循环去持续监听，否则会在 TTL 时自动将该service认为是为不可用
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				logrus.Panicf("no heartbeat from %s to registry, err=%v", serviceName, err)
			}
			time.Sleep(1 * time.Second)
		}

	}()
	logrus.WithFields(logrus.Fields{
		"serviceName": serviceName,
		"addr":        grpcAddr,
	}).Info("register to consul")
	return func() error {
		return registry.DeRegister(ctx, instanceID, serviceName)
	}, nil // 返回清洁函数给上一层调用
}
