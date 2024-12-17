package client

import (
	"context"
	"github.com/baobao233/gorder/common/discovery"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/common/genproto/stockpb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewStockGRPCClient 封装一层 NewStockServiceClient
func NewStockGRPCClient(ctx context.Context) (client stockpb.StockServiceClient, close func() error, err error) {
	// 从 consul 中 discover 可用的stockGRPC 的地址，而不是直接绑定固定的 stockGRPC 地址
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("stock.service-name"))
	if err != nil {
		return nil, func() error {
			return nil
		}, err
	}
	if grpcAddr == "" {
		logrus.Warn("empty grpc addr for stock grpc")
	}
	opts, err := grpcDialOpts(grpcAddr)
	if err != nil {
		return nil, func() error { return nil }, err
	} // 返回默认值
	// NewClient 返回一个连接，close 也是用于关闭该连接
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	return stockpb.NewStockServiceClient(conn), conn.Close, nil
}

// NewOrderGRPCClient 封装一层 NewOrderServiceClient
func NewOrderGRPCClient(ctx context.Context) (client orderpb.OrderServiceClient, close func() error, err error) {
	// 从 consul 中 discover 可用的stockGRPC 的地址，而不是直接绑定固定的 stockGRPC 地址
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("order.service-name"))
	if err != nil {
		return nil, func() error {
			return nil
		}, err
	}
	if grpcAddr == "" {
		logrus.Warn("empty grpc addr for order grpc")
	}
	opts, err := grpcDialOpts(grpcAddr)
	if err != nil {
		return nil, func() error { return nil }, err
	} // 返回默认值
	// NewClient 返回一个连接，close 也是用于关闭该连接
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	return orderpb.NewOrderServiceClient(conn), conn.Close, nil

}

func grpcDialOpts(addr string) ([]grpc.DialOption, error) {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, nil
}
