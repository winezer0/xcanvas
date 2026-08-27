package main

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// 使用Gin框架
	r := gin.Default()

	// 使用gRPC
	server := grpc.NewServer()
	_ = server

	r.Run()
}
