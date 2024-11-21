package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	
)

var client *mongo.Client

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB Connection
	var err error
	MONGOURL := "mongodb+srv://qas-app-rwuser:fSXQNZ7G9B38n4xi@quality-app.ss86w.mongodb.net/"
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(MONGOURL))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")

	// Fiber App
	app := fiber.New()
	app.Get("/", getData)
	if err := app.Listen(":6060"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Struct Definitions
type TaskDataStruct struct {
	ID             primitive.ObjectID   `json:"_id" bson:"_id"`
	Title          string               `json:"title" bson:"title"`
	Status         interface{}          `json:"status" bson:"status"`
	AssignedTo     interface{}          `json:"assignedTo" bson:"assignedTo"`
	RefBoardID     []primitive.ObjectID `json:"refBoardID" bson:"refBoardID"`
	PlannedFrom    interface{}          `json:"plannedFrom" bson:"plannedFrom"`
	PlannedTo      interface{}          `json:"plannedTo" bson:"plannedTo"`
	ActualStart    interface{}          `json:"startedOn" bson:"startedOn"`
	ActualEnd      interface{}          `json:"completedOn" bson:"completedOn"`
	RevisedStart   interface{}          `json:"revisedStartDate" bson:"revisedStartDate"`
	RevisedEnd     interface{}          `json:"revisedEndDate" bson:"revisedEndDate"`
	Participants   interface{}          `json:"participants" bson:"participants"`
	Tags           interface{}          `json:"tags" bson:"tags"`
	Role           interface{}          `json:"role" bson:"role"`
	TaskType       interface{}          `json:"type" bson:"type"`
	WorkStream     interface{}          `json:"workstream" bson:"workstream"`
}
type BoardDataStruct struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title    string             `json:"title" bson:"title"`
	Category string             `json:"category" bson:"category"`
	Tasks []TaskDataStruct      `json:"tasks" bson:"tasks"`
}
type AssignedToStruct struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FullName string             `json:"fullName" bson:"fullName"`
	Email    string             `json:"email" bson:"email"`
}
type StatusDataStruct struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Category string             `json:"category" bson:"category"`
	Status   string             `json:"status" bson:"status"`
	WorkItem string             `json:"workItem" bson:"workItem"`
}
type Mapstruct struct {
	AssignedToMap map[primitive.ObjectID]AssignedToStruct
	StatusMap map[primitive.ObjectID]StatusDataStruct
}
type RWTStruct struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`
	Name string `json:"name" bson:"name"`
	Type string `json:"__type" bson:"__type"`
}
func createMapForUSerAndStatus(assignedToData []AssignedToStruct, statusData []StatusDataStruct)Mapstruct{
	assignedToMap := make(map[primitive.ObjectID]AssignedToStruct)
	statusMap := make(map[primitive.ObjectID]StatusDataStruct)
	for _,user := range assignedToData {
		assignedToMap[user.ID] = user
	}
	for _,status := range statusData {
		statusMap[status.ID] = status
	}
	return Mapstruct{
		AssignedToMap :assignedToMap,
		StatusMap:statusMap,
	}
}
// func createMapForRWT(rwtData []RWTStruct) map[primitive.ObjectID]RWTStruct{
// 	rwtMap:= make(map[primitive.ObjectID]RWTStruct)
// 	for _,rwt:= range rwtData {
// 		rwtMap[rwt.ID] = rwt
// 	}
// 	return rwtMap
// }
func createMapForRWT(rwtData []RWTStruct)map[primitive.ObjectID]RWTStruct{
	rwtMap := make(map[primitive.ObjectID]RWTStruct)
	for _, rwt:=range rwtData {
		rwtMap[rwt.ID] = rwt
	}
	return rwtMap
}
func createBoardMap(boardData []BoardDataStruct)map[primitive.ObjectID]*BoardDataStruct{
	boardMap:= make(map[primitive.ObjectID]*BoardDataStruct)
	for board := range boardData {
		boardMap[boardData[board].ID] = &boardData[board]
	}
	
	return boardMap
}
func assignOwnerAndStatus(taskData *TaskDataStruct, userMap map[primitive.ObjectID]AssignedToStruct, statusMap map[primitive.ObjectID]StatusDataStruct){
	// fmt.Println("taskData",taskData)
	if assignedTo,ok := taskData.AssignedTo.(primitive.ObjectID);ok{
		if user,exists := userMap[assignedTo];exists{
			taskData.AssignedTo = user
		}
	}
	if status,exists := statusMap[taskData.Status.(primitive.ObjectID)];exists{
		taskData.Status = status
	}
}
func asisgnRWTvalues(taskData *TaskDataStruct,rwtMap map[primitive.ObjectID]RWTStruct){
	 if isRoleArray,ok := taskData.Role.(primitive.A);ok{
		if len(isRoleArray) > 0 {
			if roleOk,ok:=isRoleArray[0].(primitive.ObjectID);ok{
				if role,exists := rwtMap[roleOk];exists{
					taskData.Role = role
				}
			}
		}
	 }
	 if isTaskTypeArray,ok := taskData.TaskType.(primitive.A);ok{
		if len(isTaskTypeArray) > 0 {
			if taskTypeOk,ok:=isTaskTypeArray[0].(primitive.ObjectID);ok{
				if taskType,exists := rwtMap[taskTypeOk];exists{
					taskData.TaskType = taskType
				}
			}
		}
	 }
	 if isWorkStreamArray,ok := taskData.WorkStream.(primitive.A);ok{
		if len(isWorkStreamArray) > 0 {
			if workstreamok,ok:=isWorkStreamArray[0].(primitive.ObjectID);ok{
				if workstream,exists := rwtMap[workstreamok];exists{
					taskData.WorkStream = workstream
				}
			}
		}
	 }
}
func getData(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		boardData      []BoardDataStruct
		taskData       []TaskDataStruct
		assignedToData []AssignedToStruct
		statusData     []StatusDataStruct
		rwtData        []RWTStruct
	)

	// MongoDB Collections and Pipelines
	boardDB := client.Database("google-wp").Collection("kt_m_boards")
	taskDB := client.Database("google-wp").Collection("kt_t_taskLists")
	assignedToDB := client.Database("google-wp").Collection("kt_m_users")
	statusDB := client.Database("google-wp").Collection("kt_m_status")
	rwtDB := client.Database("google-wp").Collection("kt_m_types")
	// Aggregation Pipelines
	boardPipeline := mongo.Pipeline{
		{{"$match", bson.D{{"category", "custom"}, {"active", true}}}},
		{{"$project", bson.D{{"_id", 1}, {"title", 1}, {"category", 1}}}},
	}
	taskPipeline := mongo.Pipeline{
		{{"$match", bson.D{{"refBoardID", bson.D{{"$exists", true}}}, {"board", "custom"}, {"skip", false}}}},
		{{"$project", bson.D{{"_id", 1}, {"title", 1}, {"status", 1}, {"assignedTo", 1}, {"refBoardID", 1}, {"plannedFrom", 1}, {"plannedTo", 1}, {"startedOn", 1}, {"completedOn", 1}, {"revisedStartDate", 1}, {"revisedEndDate", 1}, {"participants", 1}, {"tags", 1}, {"role", 1}, {"type", 1}, {"workstream", 1}}}},
	}
	assignedToPipeline := mongo.Pipeline{
		{{"$project", bson.D{{"_id", 1}, {"fullName", 1}, {"email", 1}}}},
	}
	statusPipeline := mongo.Pipeline{
		{{"$match", bson.D{{"workItem", "Task"}}}},
		{{"$project", bson.D{{"_id", 1}, {"category", 1}, {"status", 1}, {"workItem", 1}}}},
	}
	rwtPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"$or", bson.A{
				bson.D{{"__type", "role"}},
				bson.D{{"__type", "tasktype"}},
				bson.D{{"__type", "workstream"}},
			}},
		}}},
		{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"__type", 1}}}},
	}
	// Concurrent Data Fetch
	errChan := make(chan error, 5)
	go getDBData(ctx, boardDB, boardPipeline, &boardData, errChan)
	go getDBData(ctx, taskDB, taskPipeline, &taskData, errChan)
	go getDBData(ctx, assignedToDB, assignedToPipeline, &assignedToData, errChan)
	go getDBData(ctx, statusDB, statusPipeline, &statusData, errChan)
	go getDBData(ctx, rwtDB, rwtPipeline, &rwtData, errChan)

	// Handle Errors
	for i := 0; i < 5; i++ {
		if err := <-errChan; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}
	maps := createMapForUSerAndStatus(assignedToData,statusData)
	rwtMap := createMapForRWT(rwtData)
	boardMap:= createBoardMap(boardData)
	for task := range taskData {
		assignOwnerAndStatus(&taskData[task], maps.AssignedToMap,maps.StatusMap)
		asisgnRWTvalues(&taskData[task],rwtMap)
	}
	 for _, task := range taskData {
        if board, exists := boardMap[task.RefBoardID[0]]; exists {
            board.Tasks = append(board.Tasks, task)
        }
    }
	
	// Return Data
	return c.JSON(fiber.Map{
		"data": boardData,
	})
}

// Helper Function to Fetch Data
func getDBData(ctx context.Context, coll *mongo.Collection, pipeline mongo.Pipeline, result interface{}, errChan chan error) {
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		errChan <- fmt.Errorf("aggregation error: %w", err)
		return
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, result); err != nil {
		errChan <- fmt.Errorf("cursor error: %w", err)
		return
	}
	errChan <- nil
}
