package main

import (
	"fmt"
	"log"
	"net"
	"x-tentioncrew/user-service/pkg/config"
	"x-tentioncrew/user-service/pkg/db"
	"x-tentioncrew/user-service/pkg/pb"
	"x-tentioncrew/user-service/pkg/services"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("failed at config", err.Error())
	}
	db, dbErr := db.ConnectDB(cfg)
	if dbErr != nil {
		log.Fatalln("db connection failed ", err.Error())
	}

	lis, lisErr := net.Listen("tcp", cfg.Port)
	if lisErr != nil {
		log.Fatalln("failed to listing", lisErr.Error())
	}

	fmt.Println("userService on port:", cfg.Port)
	s := services.Sever{
		DB: db,
	}
	grpcServer := grpc.NewServer()

	pb.RegisterUserSvcServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
