package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "time"
    "io"
    "errors"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/genproto/googleapis/rpc/errdetails"
    hellopb "mygrpc/pkg/grpc"
)

type myServer struct {
    hellopb.UnimplementedGreetingServiceServer
}

// 自作サービス構造体のコンストラクタを定義
func NewMyServer() *myServer {
    return &myServer{}
}

// Hello メソッドの実装
func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
    // 何か処理をしてエラーが発生した場合
    //err := status.Error(codes.Unknown, "unknown error occurred")

    stat := status.New(codes.Unknown, "unknown error occurred")
    stat, _ = stat.WithDetails(&errdetails.DebugInfo{
        Detail: "error occurred in Hello method",
    })
    err := stat.Err()

    // エラーを返す
    return nil, err

    // リクエストからnameフィールドを取り出して
    // "Hello, [名前]!"というレスポンスを返す
    // return &hellopb.HelloResponse{
    //     Message: fmt.Sprintf("Hello, %s!", req.GetName()),
    // }, nil
}

// HelloServerStream メソッドの実装
func (s *myServer) HelloServerStream(req *hellopb.HelloRequest, stream hellopb.GreetingService_HelloServerStreamServer) error {
    resCount := 5
    for i := 0; i < resCount; i++ {
        if err := stream.Send(&hellopb.HelloResponse{
            Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
        }); err != nil {
            return err
        }
        time.Sleep(time.Second * 1)
    }
    return nil
}

// HelloClientStream メソッドの実装
func (s *myServer) HelloClientStream(stream hellopb.GreetingService_HelloClientStreamServer) error {
    nameList := make([]string, 0)
    for {
        req, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            message := fmt.Sprintf("Hello, %v!", nameList)
            return stream.SendAndClose(&hellopb.HelloResponse{
                Message: message,
            })
        }
        if err != nil {
            return err
        }
        nameList = append(nameList, req.GetName())
    }
}

// HelloBiStreams メソッドの実装
func (s *myServer) HelloBiStreams(stream hellopb.GreetingService_HelloBiStreamsServer) error {
    for {
        req, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            return nil
        }
        if err != nil {
            return err
        }
        message := fmt.Sprintf("Hello, %v!", req.GetName())
        if err := stream.Send(&hellopb.HelloResponse{
            Message: message,
        }); err != nil {
            return err
        }
    }
}

func main() {
    // 1. 8080番portのLisnterを作成
    port := 8080
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        panic(err)
    }

    // 2. gRPCサーバーを作成
    s := grpc.NewServer(
        /*grpc.ChainUnaryInterceptor(
			myUnaryServerInterceptor1,
			myUnaryServerInterceptor2,
		),

        grpc.ChainStreamInterceptor(
			myStreamServerInterceptor1,
			myStreamServerInterceptor2,
		),
        grpc.UnaryInterceptor(myUnaryServerInterceptor1),
        grpc.StreamInterceptor(myStreamServerInterceptor1),*/
    )

    // 3. gRPCサーバーにGreetingServiceを登録
    hellopb.RegisterGreetingServiceServer(s, NewMyServer())

    // 4. サーバーリフレクションの設定
    reflection.Register(s)

    // 4. 作成したgRPCサーバーを、8080番ポートで稼働させる
    go func() {
        log.Printf("start gRPC server port: %v", port)
        s.Serve(listener)
    }()

    // 5.Ctrl+Cが入力されたらGraceful shutdownされるようにする
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    log.Println("stopping gRPC server...")
    s.GracefulStop()
}
