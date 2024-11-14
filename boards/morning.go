package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	
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
	Id  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title string `json:"title" bson:"title"`
	Category string `json:"category" bson:"category"`
	Tasks []TaskDetails `json:"tasks" bson:"tasks"`
}
type TaskDetails struct {
	Id  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title string `json:"title" bson:"title"`
	RefBoardID []primitive.ObjectID `json:"refBoardID,omitempty" bson:"refBoardID,omitempty"`
	AssignedTo interface{}  `json:"assignedTo" bson:"assignedTo"`
	Status interface{} `json:"status,omitempty" bson:"status,omitempty"`
}
type AssignedToDetails struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FullName string `json:"fullName" bson:"fullName"`
	Email string `json:"email" bson:"email"` 
}
type StatusDetails struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Category string `json:"category" bson:"category"`
	Status string `json:"status" bson:"status"` 
	WorkItem string `json:"workItem" bson:"workItem"` 
}
type BoardWithTask struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Tasks TaskDetails `json:"tasks" bson:"tasks`
}
func replaceAssignedTo(task *TaskDetails, assignedDetails []AssignedToDetails) {
	for _, assigned := range assignedDetails {
		if task.AssignedTo == assigned.Id {
			task.AssignedTo = assigned
			break
		}
	}
}
func replaceStatus(task *TaskDetails, statusDetails []StatusDetails){
	for _,status:=range statusDetails {
		if task.Status == status.Id {
			task.Status = status
		}
	}
}
func getData(c *fiber.Ctx)error{
	var boardDetails []BoardDetails
	var taskDetails []TaskDetails
	var assignedToDetails []AssignedToDetails
	var statusDetails []StatusDetails

	boardDBDetails := client.Database("kternai-tp0").Collection("kt_m_boards")
	taskDBDetails := client.Database("kternai-tp0").Collection("kt_t_taskLists")
	assignedToDBDetails :=  client.Database("kternai-tp0").Collection("kt_m_users")
	statusDBDetails := client.Database("kternai-tp0").Collection("kt_m_status")

	boardChan := make(chan error)
	taskChan := make(chan error)

	boardDataPipeline := mongo.Pipeline{
        {{"$match", bson.D{{"category", "custom"},{"active", true}}}},
        {{"$project", bson.D{{"_id", 1}, {"title", 1}, {"category", 1}}}},
    }

    taskDataPipeline := mongo.Pipeline{
        {{"$match", bson.D{{"refBoardID", bson.D{{"$exists", true}}}, {"board", "custom"},{"skip",false}}}},
        {{"$project", bson.D{{"_id", 1}, {"title", 1},{"refBoardID",1}, {"assignedTo",1},{"status",1}}}},
    }
	assignedToDataPipeline := mongo.Pipeline{
		{{"$project", bson.D{{"_id",1},{"fullName",1},{"email",1}}}},
	}
	statusDataPipeline := mongo.Pipeline{
		{{"$match",bson.D{{"workItem","Task"}}}},
		{{"$project", bson.D{{"_id",1},{"category",1},{"status",1},{"workItem",1}}}},
	}
	go func(){
		cursor,err := assignedToDBDetails.Aggregate(context.Background(),assignedToDataPipeline)
		defer cursor.Close(context.Background())
		if err != nil {
			fmt.Println("Error in getting in data")
		}
		for cursor.Next(context.Background()){
			var data AssignedToDetails
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("Error in decoding AssignedTo data")
			}
			assignedToDetails = append(assignedToDetails,data)
		}
	} ()
	go func(){
		fmt.Println("BOARD FUNCTION CALLED")
		cursor,err := boardDBDetails.Aggregate(context.Background(),boardDataPipeline)
		if err != nil {
			fmt.Println("Error in Retreving Data")
			boardChan <- fmt.Errorf("error in retreving data %v",err)
			return 
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()){
			var data BoardDetails
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("Error in Decoding Data from board details")
				boardChan <- fmt.Errorf("error in retreving data %v",decodedError)
				return 
			}
			boardDetails = append(boardDetails,data)
		}
		boardChan <- nil 
	} ()
	go func(){
		fmt.Println("TASK FUNCTION CALLED")
		cursor,err := taskDBDetails.Aggregate(context.Background(),taskDataPipeline)
		if err != nil {
			fmt.Println("Error in Retreving Data")
			taskChan <- fmt.Errorf("error in retreving data %v",err)
			return
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()){
			var data TaskDetails
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("Error in Decoding Data from task details")
				taskChan <- fmt.Errorf("error in retreving data %v",decodedError)
				return 
			}
			if data.AssignedTo == "" {
				data.AssignedTo = ""
			} else {
				data.AssignedTo = data.AssignedTo
			}
			taskDetails = append(taskDetails,data)
		}
		fmt.Println("Length od doc:",len(taskDetails))
		taskChan <- nil 
	} ()
	go func (){
		cursor,err := statusDBDetails.Aggregate(context.Background(),statusDataPipeline)
		if err != nil {
			fmt.Println("Error in status retreving")
		}
		for cursor.Next(context.Background()){
			var data StatusDetails
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("Error in decoding status data")
			}
			statusDetails = append(statusDetails,data)
		}
	} ()
	 if err:= <- boardChan; err!= nil {
		fmt.Println("error in channel",err)
	 }
	 if err:= <- taskChan; err!= nil {
		fmt.Println("error in channel",err)
	 }
	 for i := range taskDetails {
	  replaceAssignedTo(&taskDetails[i], assignedToDetails)
	  replaceStatus(&taskDetails[i], statusDetails)
	 }
	 for i, board := range boardDetails {
        for _, task := range taskDetails {
            // Check if the task's RefBoardID matches the current board's Id
            if len(task.RefBoardID) > 0 && board.Id == task.RefBoardID[0] {
                boardDetails[i].Tasks = append(boardDetails[i].Tasks, task)
            }
        }
    }
	return c.JSON(
		fiber.Map{
			"status":200,
			"boardDetails":boardDetails,
		},
	)
}