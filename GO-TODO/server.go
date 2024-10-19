package main

import(
	"fmt"
	"my-go-backend/db"
	"my-go-backend/redise"
	"github.com/gofiber/fiber/v2"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
 	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/gofiber/fiber/v2/middleware/cors"
	 "github.com/redis/go-redis/v9"
	 "time"
	 "encoding/json"
)
var (
	client *mongo.Client
	redisClient *redis.Client
) 

func main(){
	app := fiber.New()
	app.Use(cors.New())

	var err error
	client, err = db.DbCon()
	if err != nil {
		fmt.Println("Error in Db Connection")
	}
	defer client.Disconnect(context.Background())
	
	redisClient,err = redise.RedisConnect()
	if err != nil{
		fmt.Println("error in connecting redis")
	}
	fmt.Println("Redis connection established")
	fmt.Println("Redis", redisClient)
	fmt.Println("Db Connection made Successfully")

	app.Get("/", getData)
	app.Post("/posttodo",postData)
	app.Put("/updatetodo",updateData)
	app.Delete("/deleteTodo",deleteTodo)

	port := app.Listen(":7000")
	if port != nil {
		fmt.Println("Error in connecting the Port")
	}

}
type UserDetail struct{
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string `json:"name",bson:"name"`
	Age int `json:"age",bson:"age"`
}
func getData(c *fiber.Ctx)error{
	getDataFromRedis,getDataFromRedisError := redisClient.Get(context.Background(),"gettodo").Result()
	fmt.Println("data from redis",getDataFromRedis)
	if getDataFromRedisError != nil {
		fmt.Println("Error in redis getdata")
	}
	var redisdata []UserDetail
	unparsedDataerr := json.Unmarshal([]byte(getDataFromRedis), &redisdata)
	if unparsedDataerr != nil {
		fmt.Println("Error in parsed data")
	}
	fmt.Println("redis data in go struct",redisdata)
	if redisdata != nil {
		return c.JSON(fiber.Map{"status":200,"msg":"data from redis", "data":redisdata})
	}
	var users []UserDetail
	dbcollection := client.Database("go-backend").Collection("gettodo")
	dbcollectionData,dbCollectionErrorData := dbcollection.Find(context.Background(),bson.M{})
	if dbCollectionErrorData != nil {
		fmt.Println("Error in collection")
	}
	defer dbcollectionData.Close(context.Background())

	for dbcollectionData.Next(context.Background()){
		var user UserDetail
		decodedData := dbcollectionData.Decode(&user)
		if decodedData!= nil {
			fmt.Println("Error in Decoded data")
			return decodedData
		}
		users = append(users,user)
	}
	collectionDataError :=  dbcollectionData.Err()

	if collectionDataError !=nil {
		return collectionDataError
	}
	// fmt.Println("response data",users)
	parsedData, parsedDataError := json.Marshal(users)
	if parsedDataError != nil {
		fmt.Println("Error in Parsing Data")
	}
	cachingData,rErr := redisClient.Set(context.Background(),"gettodo",parsedData,24*time.Hour).Result()
	fmt.Println("cachedata",cachingData)
	if rErr != nil {
		fmt.Println("redis error in", rErr)
	}
	return c.JSON(fiber.Map{"status":200,"msg":"data from database", "data":users})
}


type PostDataDetails struct{
	Name string `json:"name" bson:"name"`
	Age int `json:"age" bson:"age"`
}

func postData(c *fiber.Ctx)error{
	reqData := new(PostDataDetails)
	parsedDataErr := c.BodyParser(reqData)
	if parsedDataErr != nil {
		fmt.Println("Error in Parsing Data %+v",parsedDataErr)
		return nil
	}
	fmt.Println("parsed value is %+v\n",reqData)

	addingDataToDb := client.Database("go-backend").Collection("gettodo")
	cursor,err := addingDataToDb.InsertOne(context.Background(),reqData)
	if err != nil {
		fmt.Println("Error in inserting data")
		return nil
	}
	defer client.Disconnect(context.Background())
	return c.JSON(fiber.Map{"status":200,"data":cursor})
}
type updateDataDetails struct {
	ID string `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
	Age int `json:"age" bson:"age`
}
func updateData(c *fiber.Ctx)error{
	fmt.Println("i am inside update")
	reqData := new(updateDataDetails)

	parsedError := c.BodyParser(reqData)
	if parsedError != nil {
		fmt.Println("Error in parsing data %v\n", parsedError)
		return nil
	}


	objID, err := primitive.ObjectIDFromHex(reqData.ID)
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format")
    }
	dbdetails := client.Database("go-backend").Collection("gettodo")
	filter := bson.M{ "_id":objID}
	
	update := bson.M{"$set":bson.M{
		"name":reqData.Name,
		"age":reqData.Age,
	}}

	dbconnection,dberror := dbdetails.UpdateOne(context.Background(), filter, update)
	
	if dberror != nil {
		fmt.Println("Error in db connecting")
		return dberror
	}

	if dbconnection.MatchedCount == 0 {
		return c.JSON(fiber.Map{"status":400, "msg":"there is no document with id you send "})
	}

	return c.JSON(fiber.Map{"status":200,"msg": "updated successfully","data":reqData, "updatedata": dbconnection,})
}

func deleteTodo(c *fiber.Ctx)error{
	 queryvalue := c.Query("projectID")
	fmt.Println("query value is",queryvalue)
	var header = c.Get("authorization")
	fmt.Println("header value is",header)
	return c.JSON(fiber.Map{"queryvalue":queryvalue,"header":header})
}