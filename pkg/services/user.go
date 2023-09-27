package services

import (
	"x-tentioncrew/user-service/pkg/pb"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Sever struct {
	DB      *gorm.DB
	RedisDB *redis.Client
	pb.UnimplementedUserSvcServer
}

func CreatUserUser() {

}
