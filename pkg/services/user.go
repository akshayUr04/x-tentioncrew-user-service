package services

import (
	"x-tentioncrew/user-service/pkg/pb"

	"gorm.io/gorm"
)

type Sever struct {
	DB *gorm.DB
	pb.UnimplementedUserSvcServer
}

func CreatUserUser() {

}
