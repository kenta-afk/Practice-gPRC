package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    chatpb "mygrpc/pkg/grpc"

)

var (
    scanner *bufio.Scanner
    chatClient chatpb.ChatServiceClient
)

func main() {
    fmt.Println("start gRPC Client.")

    scanner = bufio.NewScanner(os.Stdin)

    address := "localhost:8080"
    conn, err := grpc.Dial(
        address,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        log.Fatal("Connection failed.")
        return
    }
    defer conn.Close()

    chatClient = chatpb.NewChatServiceClient(conn)

    for {
        fmt.Println("1: Start Chat")
        fmt.Println("2: Exit")
        fmt.Print("Please enter > ")

        scanner.Scan()
        in := scanner.Text()

        switch in {
        case "1":
            startChat()

        case "2":
            fmt.Println("bye.")
            return
        }
    }
}

func startChat() {
    stream, err := chatClient.Chat(context.Background())
    if err != nil {
        log.Fatalf("Failed to start chat: %v", err)
    }

    go func() {
        for {
            msg, err := stream.Recv()
            if err != nil {
                log.Fatalf("Failed to receive message: %v", err)
            }
            fmt.Printf("%s: %s\n", msg.User, msg.Message)
        }
    }()

    fmt.Println("Enter your username:")
    scanner.Scan()
    username := scanner.Text()

    fmt.Println("Start chatting! Type your messages below:")

    for scanner.Scan() {
        text := scanner.Text()
        if text == "exit" {
            break
        }
        err := stream.Send(&chatpb.ChatMessage{
            User:    username,
            Message: text,
        })
        if err != nil {
            log.Fatalf("Failed to send message: %v", err)
        }
    }
    stream.CloseSend()
}