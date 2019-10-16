package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	api "github.com/hpifu/go-godtoken/api"
	"google.golang.org/grpc"
)

var AppVersion = "unknown"

func main() {
	version := flag.Bool("v", false, "print current version")
	address := flag.String("h", "127.0.0.1:7060", "address")
	flag.Parse()
	if *version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("dial failed. err: [%v]\n", err)
		return
	}
	defer conn.Close()

	client := api.NewServiceClient(conn)

	res, err := client.GetToken(context.Background(), &api.GetTokenReq{Rid: "1234567"})
	fmt.Println(res)
}
