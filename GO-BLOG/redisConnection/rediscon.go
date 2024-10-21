package redisConnection

import (
	"github.com/redis/go-redis/v9"
	"fmt"
	"context"
)
var redisClient *redis.Client

func RedisConnection()(*redis.Client,error){
	REDISURL := "localhost:9001"
	client := redis.NewClient(&redis.Options{
        Addr:	  REDISURL,
        Password: "", // No password set
        DB:		  0,  // Use default DB
        Protocol: 2,  // Connection protocol
    })

	redisC,redisCError := client.Ping(context.Background()).Result()
	if redisCError!= nil {
		fmt.Println("Error in Redis")
		return nil,redisCError
	}
	fmt.Println("REdis connected Successfully",redisC)
	return client,nil
}