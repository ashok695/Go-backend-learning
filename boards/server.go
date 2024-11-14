package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
)
var client *mongo.Client
func main(){
	var err error
	MONGODB_URL := "mongodb+srv://qas-app-rwuser:fSXQNZ7G9B38n4xi@quality-app.ss86w.mongodb.net/"
	clientOptions := options.Client().ApplyURI(MONGODB_URL)
	client,err = mongo.Connect(context.Background(),clientOptions)
	if err != nil {
		fmt.Println("Error in Connecting the DB")
	}
	pingErr := client.Ping(context.Background(),nil)
	if pingErr != nil {
		fmt.Println("Error in pinging the data")
	}
	fmt.Println("DB connection made Success")
	fmt.Println("DB connection made Success",client)
	defer client.Disconnect(context.Background())
	app:= fiber.New()
	app.Get("/",getData)
	appListenError := app.Listen(":8000")
	if appListenError != nil {
		fmt.Println("Error in port connecting")
	}
}
type BoardDetails struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title    string             `json:"title" bson:"title"`
	Category string             `json:"category" bson:"category"`
}

type TaskDetails struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title string             `json:"title" bson:"title"`
	Board string             `json:"board" bson:"board"`
}

func getData(c *fiber.Ctx) error {
	var boardDetails []BoardDetails
	var taskDetails []TaskDetails

	collectionData := client.Database("google-pr").Collection("kt_m_boards")
	taskData := client.Database("google-pr").Collection("kt_t_taskLists")

	// Pipelines for aggregation
	boardPipeline := mongo.Pipeline{
		{{"$match", bson.D{{"category", "custom"}}}},
		{{"$project", bson.D{{"_id", 1}, {"title", 1}, {"category", 1}}}},
	}

	taskPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"refBoardID", bson.M{"$exists": true}},
			{"board", "custom"},
		}}},
	}

	// Using channels to run aggregation concurrently
	boardChan := make(chan error)
	taskChan := make(chan error)

	// Run board aggregation in a goroutine
	go func() {
		cursor, err := collectionData.Aggregate(context.Background(), boardPipeline)
		if err != nil {
			boardChan <- fmt.Errorf("error in board data aggregation: %v", err)
			return
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var board BoardDetails
			if err := cursor.Decode(&board); err != nil {
				boardChan <- fmt.Errorf("error decoding board data: %v", err)
				return
			}
			boardDetails = append(boardDetails, board)
		}
		boardChan <- nil
	}()

	// Run task aggregation in a goroutine
	go func() {
		cursor, err := taskData.Aggregate(context.Background(), taskPipeline)
		if err != nil {
			taskChan <- fmt.Errorf("error in task data aggregation: %v", err)
			return
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var task TaskDetails
			if err := cursor.Decode(&task); err != nil {
				taskChan <- fmt.Errorf("error decoding task data: %v", err)
				return
			}
			taskDetails = append(taskDetails, task)
		}
		taskChan <- nil
	}()

	// Wait for both goroutines to complete
	if err := <-boardChan; err != nil {
		log.Println(err)
	}
	if err := <-taskChan; err != nil {
		log.Println(err)
	}

	// Respond with JSON
	return c.JSON(fiber.Map{
		"status":      200,
		"msg":         "data from server",
		"boards":      boardDetails,
		"tasks":       taskDetails,
		"taskslength": len(taskDetails),
	})
}