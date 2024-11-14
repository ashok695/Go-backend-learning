package db

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"context"
    "fmt"
)
var (client *mongo.Client)
func DbCon()(*mongo.Client,error){
	fmt.Println("hello from dbconnection")
	MONGODBURL := "mongodb://localhost:27017/"
	clientOptions := options.Client().ApplyURI(MONGODBURL)
	client, err := mongo.Connect(context.Background(),clientOptions)

	if err!=nil {
		return nil,err
	}

	pingErr := client.Ping(context.Background(),nil)
	if pingErr !=nil {
		return nil, pingErr
	}

	fmt.Println("DB CONNECTED SUCCESSFULLY")
	return client,nil
}