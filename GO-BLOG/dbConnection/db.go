package dbConnection

import(
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"context"
	"fmt"
)

var dbClient *mongo.Client

func MongoDBConnection()(*mongo.Client, error){
	MONGODB_URL := "mongodb://localhost:27017/"
	clientOptions := options.Client().ApplyURI(MONGODB_URL)
	client, clientError := mongo.Connect(context.Background(),clientOptions)
	if clientError != nil {
		fmt.Println("Error in Connecting Db")
		return nil, clientError
	}
	pingError := client.Ping(context.Background(),nil)
	if pingError != nil {
		fmt.Println("Error in pining the database")
		return nil,pingError
	}
	
	fmt.Println("DB connection Successfully initiated")
	return client,nil
}