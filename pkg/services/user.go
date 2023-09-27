package services

import (
	"context"
	"net/http"
	"x-tentioncrew/user-service/pkg/pb"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Server struct {
	DB      *gorm.DB
	RedisDB *redis.Client
	pb.UnimplementedUserSvcServer
}

func (s *Server) CreatUser(ctx context.Context, req *pb.CreatUserRequest) (*pb.CommonResponse, error) {
	creatUser := `INSERT INTO users(name,house_name,city,email,phone_number)VALUES($1,$2,$3,$4,$5)`
	if err := s.DB.Exec(creatUser, req.Name, req.HouseName, req.City, req.Email, req.PhoneNumber).Error; err != nil {
		return &pb.CommonResponse{
			Status: http.StatusBadRequest,
			Error:  "cant create user",
		}, err
	}
	return &pb.CommonResponse{
		Status: http.StatusCreated,
		Error:  "",
	}, nil
}
