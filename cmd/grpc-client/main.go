package main

import (
	"github.com/rpcpool/yellowstone-grpc/examples/golang/config"
	grpc "github.com/rpcpool/yellowstone-grpc/examples/golang/grpc"
)

func main() {
	config := config.ReadConfig("config/config.yml")

	conn := grpc.Connect(config.Address, config.InsecureConnection)
	defer conn.Close()

	grpc.Subscribe(conn, config)
}
