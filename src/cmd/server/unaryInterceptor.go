package main

import (
    "context"
    "log"

    "google.golang.org/grpc"
)

func myUnaryServerInterceptor1(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    log.Println("[pre] my unary server interceptor 1: ", info.FullMethod) // ハンドラの前に割り込ませる前処理
    res, err := handler(ctx, req) // 本来の処理
    log.Println("[post] my unary server interceptor 1: ", info.FullMethod) // ハンドラの後に割り込ませる後処理
    return res, err
}



func myUnaryServerInterceptor2(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("[pre] my unary server interceptor 2: ", info.FullMethod, req)
	res, err := handler(ctx, req) // 本来の処理
	log.Println("[post] my unary server interceptor 2: ", res)
	return res, err
}