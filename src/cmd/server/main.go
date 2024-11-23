package main

import (
    "database/sql"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "time"
    "io"
    

    _ "github.com/go-sql-driver/mysql"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    chatpb "mygrpc/pkg/grpc"
)

type chatServer struct {
    chatpb.UnimplementedChatServiceServer
    db *sql.DB
}

func NewChatServer(db *sql.DB) *chatServer {
    return &chatServer{db: db}
}

func (s *chatServer) Chat(stream chatpb.ChatService_ChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }
        log.Printf("Received message from %s: %s", msg.User, msg.Message)

        // メッセージをデータベースに保存
        _, err = s.db.Exec("INSERT INTO messages (user, message, timestamp) VALUES (?, ?, ?)", msg.User, msg.Message, time.Now().Unix())
        if err != nil {
            return err
        }

        if err := stream.Send(msg); err != nil {
            return err
        }
    }
}

func main() {
    // データベース接続
    dsn := "chatuser:chatpassword@tcp(127.0.0.1:3306)/chatdb"
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // テーブル作成
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
        id INT AUTO_INCREMENT PRIMARY KEY,
        user VARCHAR(255),
        message TEXT,
        timestamp BIGINT
    )`)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }

    // 8080番portのLisnterを作成
    port := 8080
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        panic(err)
    }

    // gRPCサーバーを作成
    s := grpc.NewServer()
    chatpb.RegisterChatServiceServer(s, NewChatServer(db))
    reflection.Register(s)

    // 作成したgRPCサーバーを、8080番ポートで稼働させる
    go func() {
        log.Printf("start gRPC server port: %v", port)
        s.Serve(listener)
    }()

    // Ctrl+Cが入力されたらGraceful shutdownされるようにする
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    log.Println("stopping gRPC server...")
    s.GracefulStop()
}
