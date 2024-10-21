package main

import(
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go-blog/dbConnection" 
	"go-blog/redisConnection" 
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
 	"go.mongodb.org/mongo-driver/bson/primitive"
	"context"
	"github.com/redis/go-redis/v9"
	"encoding/json"
	
	"github.com/golang-jwt/jwt/v5"
	"time"
)
var (
	dbClient *mongo.Client
	cacheClient *redis.Client
	jwt_secret_key = "ashok"
 )
func main(){
	app := fiber.New()
	
	var err error

	dbClient, err = dbConnection.MongoDBConnection()
	if err != nil {
		fmt.Println("error in connecting database")
	}
	fmt.Println("db connected successfully %+v /n",dbClient)
	//redis Connection
	cacheClient, err = redisConnection.RedisConnection()
	if err != nil {
		fmt.Println("error in connecting redis",err)
	}
	fmt.Println("REDIS CONNECTION ESTABLISHED",cacheClient)
	app.Get("/",getBlog)
	app.Post("/verifyUser",verifyUser)
	defer dbClient.Disconnect(context.Background())
	portServer := app.Listen(":9000")
	if portServer != nil {
		fmt.Println("Error in connecting Port")
	}
	fmt.Println("Server is ready")
}
type photoDetail struct {
	Id  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	AlbumID int `json:"albumId" bson:"albumId"`
	Title string `json:"title" bson:"title"`
	Url string `json:"url" bson:"url"`
	ThumbnailUrl string `json:"thumbnailUrl" bson:"thumbnailUrl"`
}
func getBlog(c *fiber.Ctx)error{
	getPhotoDataFromRedis,getPhotoDataFromRedisError := cacheClient.Get(context.Background(),"getphotos").Result()
	if getPhotoDataFromRedisError != nil {
		fmt.Println("NO DATA IN REDIS")
	}
	var dataFromRedis []photoDetail
	unParsingDataError := json.Unmarshal([]byte(getPhotoDataFromRedis),&dataFromRedis)
	if unParsingDataError != nil {
		fmt.Println("ERROR IN UNPARSING DATA")
	}
	if dataFromRedis != nil {
		return c.JSON(fiber.Map{"data":dataFromRedis,"status":200,"msg":"data from redis"})
	}
	var photos []photoDetail 
	collectionData := dbClient.Database("admin").Collection("photos")
	dbData,dbDataError := collectionData.Find(context.Background(),bson.M{})
	if dbDataError != nil {
		fmt.Println("Error in retreving Data")
	}
	defer dbData.Close(context.Background())
	for	dbData.Next(context.Background()){
		var photo photoDetail
		decodedDataError := dbData.Decode(&photo)
		if decodedDataError != nil {
			fmt.Println("error in decoding")
		}
		photos = append(photos,photo)
	}
	collectonDataErr := dbData.Err()
	if collectonDataErr !=nil {
		fmt.Println("error in collection Data error")
	}
	parsedData,parsedDataError := json.Marshal(photos)
	if parsedDataError != nil {
		fmt.Println("Error in parsing Data",parsedDataError)
	}
	setRedisData,setRedisDataError := cacheClient.Set(context.Background(),"getphotos",parsedData,0).Result()
	if setRedisDataError != nil{
		fmt.Println("Error in set data in redis")
	}
	fmt.Println("settdata",setRedisData)
	fmt.Println("i am inisde go route")
	return c.JSON(fiber.Map{"status":200,"msg":"data from db","data":photos})
}
type loginStruct struct{
	UserName string `json:"userName" bson:"userName"`
	Password string `json:"password" bson:"password"`
}
func verifyUser(c *fiber.Ctx)error{
	reqUser := new(loginStruct)
	reqUserParseError := c.BodyParser(reqUser)
	if reqUserParseError != nil {
		fmt.Println("ERROR IN PARSING DATA")
	}
	 userName := "ashok"
	 password := "123"
	fmt.Println("REQUEST DATA IS ",reqUser)
	if reqUser.UserName != userName || reqUser.Password != password {
		return c.JSON(fiber.Map{"status":401,"msg":"username || password is not correct "})
	}

	tokenExpires := time.Hour *24
	claims:= jwt.MapClaims{
		"username":reqUser.UserName, 
		"password":reqUser.Password,
		"exp":time.Now().Add(tokenExpires).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	resToken,err := token.SignedString([]byte(jwt_secret_key))
	if err != nil {
		fmt.Println("ERROR IN WHILE GENERATING TOKEN")
	}
	fmt.Println("TOKEN GENERATED")
	fmt.Println("TOKEN GENERATED",resToken)
	return c.JSON(fiber.Map{"status":200,"token":resToken})
}