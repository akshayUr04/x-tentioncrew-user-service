package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"x-tentioncrew/user-service/pkg/models"
	"x-tentioncrew/user-service/pkg/pb"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	cacheDuration = 2 * time.Hour
)

type Server struct {
	DB      *gorm.DB
	RedisDB *redis.Client
	pb.UnimplementedUserSvcServer
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreatUserRequest) (*pb.CommonResponse, error) {
	var user models.User
	// Store Data in Psql DB
	creatUser := `INSERT INTO users(name,house_name,city,email,phone_number)VALUES($1,$2,$3,$4,$5) RETURNING *`
	if err := s.DB.Raw(creatUser, req.Name, req.HouseName, req.City, req.Email, req.PhoneNumber).Scan(&user).Error; err != nil {
		return &pb.CommonResponse{
			Status: http.StatusBadRequest,
			Error:  "cant create user",
		}, err
	}

	// Caching the user details
	redisKey := createRedisId(user.Id)
	jsonData, _ := json.Marshal(user)
	err := s.cacheData(ctx, redisKey, jsonData)
	if err != nil {
		return &pb.CommonResponse{
			Status: http.StatusNotImplemented,
			Error:  "cant cache data",
		}, err
	}

	return &pb.CommonResponse{
		Status: http.StatusCreated,
		Error:  "",
	}, nil
}

func (s *Server) GetUserById(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponce, error) {
	// var user *pb.GetUserResponce
	var userDetails models.User
	//Check if the user is present in cache if yes return it
	redisKey := createRedisId(uint(req.Id))
	val, err := s.RedisDB.Get(ctx, redisKey).Bytes()
	if err == nil {
		fmt.Println("from redis")
		redisDetails := toJson(val)
		fmt.Println(userDetails.Name)
		return &pb.GetUserResponce{
			Name:        redisDetails.Name,
			HouseName:   redisDetails.HouseName,
			City:        redisDetails.City,
			Email:       redisDetails.Email,
			PhoneNumber: int64(redisDetails.PhoneNumber),
		}, nil
	} else if err != redis.Nil {
		fmt.Println("redis err", err.Error())
		return &pb.GetUserResponce{Error: "error while accessing cache data"}, err
	}

	// if the user not present in the cache get the details from the data base and cache the details
	query := `SELECT id,name,email,house_name,city,phone_number FROM users WHERE id = $1`
	dberr := s.DB.Raw(query, uint(req.Id)).Scan(&userDetails).Error
	if dberr != nil {
		fmt.Println("Database error:", dberr)
		return &pb.GetUserResponce{Error: "error while accessing db data"}, dberr
	}

	if userDetails.Id == 0 {
		return &pb.GetUserResponce{Error: "no user found"}, fmt.Errorf("no user found")
	}

	//cache the data
	jsonData, _ := json.Marshal(userDetails)
	err = s.cacheData(ctx, redisKey, jsonData)
	if err != nil {
		return &pb.GetUserResponce{}, err
	}

	return &pb.GetUserResponce{
		Name:        userDetails.Name,
		HouseName:   userDetails.HouseName,
		City:        userDetails.City,
		Email:       userDetails.Email,
		PhoneNumber: int64(userDetails.PhoneNumber),
	}, nil

}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.CommonResponse, error) {
	var userDetails models.User
	//update the user data in DB
	updatequery := `UPDATE users SET name=$1, house_name=$2, city=$3, email=$4, phone_number=$5 WHERE id = $6 
				   RETURNING id, name, house_name, city, email, phone_number`
	err := s.DB.Raw(updatequery, req.Name, req.HouseName, req.City, req.Email, req.PhoneNumber, req.Id).Scan(&userDetails).Error
	if err != nil {
		fmt.Println("cant update user in psql", err)
		return &pb.CommonResponse{Error: "cant update user", Status: http.StatusBadRequest}, err
	}

	if userDetails.Id == 0 {
		return &pb.CommonResponse{Error: "cant update user", Status: http.StatusBadRequest}, fmt.Errorf("no such user found")
	}

	// update cache if exists
	redisKey := createRedisId(uint(req.Id))
	_, rediErr := s.RedisDB.Get(ctx, redisKey).Bytes()
	if rediErr == redis.Nil {
		return &pb.CommonResponse{Status: http.StatusOK}, nil
	} else if rediErr != nil {
		return &pb.CommonResponse{Error: "error while accessing cache data"}, err
	} else {
		jsonData, _ := json.Marshal(userDetails)
		err := s.cacheData(ctx, redisKey, jsonData)
		if err != nil {
			return &pb.CommonResponse{Error: "error while accessing cache data"}, err
		}
	}
	return &pb.CommonResponse{Status: http.StatusOK}, nil
}

func createRedisId(Id uint) string {
	return fmt.Sprintf("userId:%v", Id)
}

func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.CommonResponse, error) {
	deleteQry := `DELETE FROM users WHERE id = $1`

	if err := s.DB.Exec(deleteQry, uint(req.Id)).Error; err != nil {
		return &pb.CommonResponse{Status: http.StatusBadRequest, Error: "cant delete user"}, err
	}

	// remove the data from redis
	redisKey := createRedisId(uint(req.Id))
	err := s.RedisDB.Del(ctx, redisKey).Err()
	if err != nil {
		return &pb.CommonResponse{Status: http.StatusBadRequest, Error: "cant delete user from redis"}, err
	}
	return &pb.CommonResponse{Status: http.StatusOK, Error: ""}, err
}

func (s *Server) GetAllUserData(ctx context.Context, req *pb.GetAllUserDataReq) (*pb.GetAllUserDataResult, error) {
	var (
		name  []string
		count int
	)
	getCount := `SELECT COUNT(id) AS count FROM users`
	getNames := `SELECT name FROM users`
	if err := s.DB.Raw(getCount).Scan(&count).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Raw(getNames).Scan(&name).Error; err != nil {
		return nil, err
	}
	return &pb.GetAllUserDataResult{Name: name}, nil

}

func toJson(val []byte) models.User {
	user := models.User{}
	err := json.Unmarshal(val, &user)
	if err != nil {
		panic(err)
	}
	return user
}

func (s *Server) cacheData(ctx context.Context, key string, value []byte) error {
	err := s.RedisDB.Set(ctx, key, value, cacheDuration).Err()
	if err != nil {
		fmt.Println("error in caheing", err)
		return err
	}
	return nil
}
