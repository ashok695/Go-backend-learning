package redise

import(
	"fmt"
	"context"
	"github.com/redis/go-redis/v9"
)
var redisClient *redis.Client
func RedisConnect()(*redis.Client,error){
	REDISURL := "localhost:9001"
	redisClient := redis.NewClient(&redis.Options{
        Addr:	 REDISURL,
        Password: "", // No password set
        DB:		  0,  // Use default DB
        Protocol: 2,  // Connection protocol
    })
	redisc,redisError := redisClient.Ping(context.Background()).Result()
	if redisError !=nil {
		fmt.Println("Error in connecting Redis")
		return nil,redisError
	}
	fmt.Println("redis connection successfuly",redisc)
	return redisClient,nil
}