package main

import(
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"context"
	"time"
	"net/http"
    _ "net/http/pprof" 
	"log"
	// "reflect"
	"sort"
)
var client *mongo.Client
func createMapForUserAndStatus(userData []AssignedToStruct, statusData []StatusStruct)Maps{
	userMap := make(map[primitive.ObjectID]AssignedToStruct)
	statusMap := make(map[primitive.ObjectID]StatusStruct)
	for _,user := range userData {
		userMap[user.ID] = user
	}
	for _,status := range statusData {
		statusMap[status.ID] = status
	}
	return Maps{
		UserMap : userMap,
		StatusMap : statusMap,
	}
}
func createMapForRWT(rwtData []RWTStruct) map[primitive.ObjectID]RWTStruct{
	rwtMap:= make(map[primitive.ObjectID]RWTStruct)
	for _,rwt:= range rwtData {
		rwtMap[rwt.ID] = rwt
	}
	return rwtMap
}
func assignOwnerAndStatus(task *TaskStruct, userMap map[primitive.ObjectID]AssignedToStruct, statusMap map[primitive.ObjectID]StatusStruct){
	// fmt.Println("task is ",task)
	if taskAssignedTo,ok := task.AssignedTo.(primitive.ObjectID);ok {
		if user,exists:= userMap[taskAssignedTo];exists{
			task.AssignedTo = user
		}
	}
	if status,exists := statusMap[task.Status.(primitive.ObjectID)];exists  && task.Status != status {
		task.Status = status
	}
}
func FindStartVariance(task *TaskStruct) {
	var actualStart time.Time
	if actualStartVal, ok := task.ActualStart.(primitive.DateTime); ok {
		actualStart = actualStartVal.Time()
	} 
	if task.PlannedFrom.IsZero() {
		task.StartVariance = Variance{
			Type:      "na",
			Days:      0,
			Parameter: "na",
			Message:   "Invalid date format",
		}
		return
	}
	plannedFrom := task.PlannedFrom
	if actualStart.IsZero() {
		task.StartVariance = Variance{
			Type:      "na",
			Days:      0,
			Parameter: "na",
			Message:   "Not Yet Started",
		}
		return
	}
	// Calculate the variance in days
	days := findDifference(plannedFrom, actualStart)
	if days < 0 {
		task.StartVariance = Variance{
			Type:      "Started",
			Days:      -days,
			Parameter: "ahead",
			Message:   fmt.Sprintf("%d days Ahead", -days),
		}
	} else if days > 0 {
		task.StartVariance = Variance{
			Type:      "Started",
			Days:      days,
			Parameter: "delayed",
			Message:   fmt.Sprintf("%d days Delayed", days),
		}
	} else {
		task.StartVariance = Variance{
			Type:      "Started",
			Days:      0,
			Parameter: "on time",
			Message:   "Started on time",
		}
	}
}
func findDifference(Date1 time.Time,Date2 time.Time)int{
	return int(Date2.Sub(Date1).Hours() / 24) // Difference in days

}
func EndVariance(task *TaskStruct){
	var actualEnd time.Time
	if actualEndCheck,ok := task.ActualEnd.(primitive.DateTime);ok{
		actualEnd = actualEndCheck.Time()
	}
	if task.PlannedTo.IsZero(){
		task.EndVariance = Variance{
			Type:      "na",
			Days:      0,
			Parameter: "na",
			Message:   "Invalid date format",
		}
		return 
	}
	if actualEnd.IsZero(){
		task.EndVariance = Variance{
			Type:      "na",
			Days:      0,
			Parameter: "na",
			Message:   "Not Yet Completed",
		}
		return 
	}
	daysDifference := findDifference(task.PlannedTo,actualEnd)
	if daysDifference > 0 {
		task.EndVariance = Variance{
			Type:      "Completed",
			Days:      daysDifference,
			Parameter: "delayed",
			Message:   fmt.Sprintf("%d days Ahead", daysDifference),
		}
	} else if daysDifference < 0 {
		task.EndVariance = Variance{
			Type:      "Completed",
			Days:      daysDifference,
			Parameter: "ahead",
			Message:   fmt.Sprintf("%d days ahead", daysDifference),
		}
	} else {
		task.EndVariance = Variance{
			Type:      "Completed",
			Days:      daysDifference,
			Parameter: "ahead",
			Message:   fmt.Sprintf("%d days Delayed", daysDifference),
		}
	}
}
func AssignRWTalue(task *TaskStruct,mapsForRWT map[primitive.ObjectID]RWTStruct){
	// fmt.Println("type of ", reflect.TypeOf(task.TaskType))
		if taskSlice,ok := task.TaskType.(primitive.A);ok{
			if len(taskSlice) > 0 {
				if  tasktype,ok := taskSlice[0].(primitive.ObjectID);ok {
					
				if hi,exists := mapsForRWT[tasktype];exists{
					
					task.TaskType = hi
				}
				}
			}
		}
		if taskSlice,ok := task.Role.(primitive.A);ok{
			if len(taskSlice) > 0 {
				if  tasktype,ok := taskSlice[0].(primitive.ObjectID);ok {
					
				if hi,exists := mapsForRWT[tasktype];exists{
					
					task.TaskType = hi
				}
				}
			}
		}
		if taskSlice,ok := task.Workstream.(primitive.A);ok{
			if len(taskSlice) > 0 {
				if  tasktype,ok := taskSlice[0].(primitive.ObjectID);ok {
					
				if hi,exists := mapsForRWT[tasktype];exists{
					
					task.TaskType = hi
				}
				}
			}
		}
}
func main(){
	ctx:= context.Background()
	var err error
	app:= fiber.New()
	MONGOURL:="mongodb+srv://qas-app-rwuser:fSXQNZ7G9B38n4xi@quality-app.ss86w.mongodb.net/"
	options:=options.Client().ApplyURI(MONGOURL)
	client,err = mongo.Connect(ctx,options)
	if err != nil {
		fmt.Println("Error in connecting Db")
	}
	pingErr := client.Ping(ctx,nil)
	if pingErr != nil {
		fmt.Println("Error in pinging the Database")
	}
	fmt.Println("DB CONNECTION SUC")
	defer client.Disconnect(ctx)
	app.Get("/",getData)
	appPort := app.Listen(":6060")
	if appPort != nil {
		fmt.Println("Error in Port")
	}
}
type SubPhaseStruct struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SubPhaseName string `json:"subPhaseName" bson:"subPhaseName"`
	Type string `json:"__type" bson:"__type"`
	OrderID string `json:"orderID" bson:"orderID"`
}
type TaskStruct struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title string `json:"title" bson:"title"`
	Status interface{} `json:"status" bson:"status"`
	AssignedTo interface {} `json:"assignedTo" bson:"assignedTo"`
	OrderID string `json:"orderID" bson:"orderID"`
	PlannedFrom time.Time `json:"plannedFrom" bson:"plannedFrom"`
	PlannedTo time.Time   `json:"plannedTo" bson:"plannedTo"`
	ActualStart interface{} `json:"startedOn" bson:"startedOn"`
	ActualEnd interface{} `json:"completedOn" bson:"completedOn"`
	StartVariance Variance `json:"startVariance" bson:"startVariance"`
	EndVariance Variance `json:"endVariance" bson:"endVariance"`
	Priority interface {} `json:"priority" bson:"priority"`
	Role interface {} `json:"role" bson:"role"`
	Workstream interface {} `json:"workstream" bson:"workstream"`
	TaskType interface {} `json:"type" bson:"type"`
}
 type StatusStruct struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Category string `json:"category" bson:"category"`
	Status string `json:"status" bson:"status"`
}
type AssignedToStruct struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FullName string `json:"fullName" bson:"fullName"`
	Email string `json:"email" bson:"email"`
}
type Maps struct {
	UserMap map[primitive.ObjectID]AssignedToStruct
	StatusMap map[primitive.ObjectID]StatusStruct
}
type Variance struct {
	Days int
	Message string
	Parameter string
	Type string
}
type RWTStruct struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`
	Name string `json:"name" bson:"name"`
	Type string `json:"__type" bson:"__type"`
}
func getData(c *fiber.Ctx) error {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil)) // Start the pprof server
	}()
	ctx := context.Background()

	// Define slices for data
	var subPhaseData []SubPhaseStruct
	var taskData []TaskStruct
	var userData []AssignedToStruct
	var statusData []StatusStruct
	var rwtData []RWTStruct

	// Define collections
	subPhaseDBDetails := client.Database("kaartechnologies-mql").Collection("kt_m_subphases")
	taskDBDetails := client.Database("kaartechnologies-mql").Collection("kt_t_taskLists")
	userDBDetails := client.Database("kaartechnologies-mql").Collection("kt_m_users")
	statusDBDetails := client.Database("kaartechnologies-mql").Collection("kt_m_status")
	rwtDBDetails := client.Database("kaartechnologies-mql").Collection("kt_m_types")

	// Pipelines
	subPhasePipeline := mongo.Pipeline{
		{{"$project", bson.D{{"_id", 1}, {"subPhaseName", 1}, {"__type", 1}, {"orderID", 1}}}},
	}
	taskPipeline := mongo.Pipeline{
		{{"$match", bson.D{{"refBoardID", bson.D{{"$exists", false}}}, {"skip", false}}}},
		{{"$project", bson.D{{"_id", 1}, {"title", 1}, {"status", 1}, {"assignedTo", 1}, {"orderID", 1}, {"plannedFrom", 1}, {"plannedTo", 1}, {"startedOn", 1}, {"completedOn", 1}, {"priority", 1}, {"role", 1}, {"workstream", 1}, {"type", 1}}}},
	}
	userPipeline := mongo.Pipeline{
		{{"$project", bson.D{{"_id", 1}, {"fullName", 1}, {"email", 1}}}},
	}
	statusPipeline := mongo.Pipeline{
		{{"$match", bson.D{{"workItem", "Task"}}}},
		{{"$project", bson.D{{"_id", 1}, {"category", 1}, {"status", 1}}}},
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

	// Concurrently fetch data
	errChan := make(chan error, 5)

	go fetchData(ctx, subPhaseDBDetails, subPhasePipeline, &subPhaseData, errChan)
	go fetchData(ctx, taskDBDetails, taskPipeline, &taskData, errChan)
	go fetchData(ctx, userDBDetails, userPipeline, &userData, errChan)
	go fetchData(ctx, statusDBDetails, statusPipeline, &statusData, errChan)
	go fetchData(ctx, rwtDBDetails, rwtPipeline, &rwtData, errChan)

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		if err := <-errChan; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	// Create maps for processing
	maps := createMapForUserAndStatus(userData, statusData)
	rwtMap := createMapForRWT(rwtData)

	// Process tasks
	for i := range taskData {
		assignOwnerAndStatus(&taskData[i], maps.UserMap, maps.StatusMap)
		FindStartVariance(&taskData[i])
		EndVariance(&taskData[i])
		AssignRWTalue(&taskData[i], rwtMap)
	}
	sort.Slice(taskData, func(i, j int) bool {
		return taskData[i].OrderID < taskData[j].OrderID
	})

	// Send response
	return c.JSON(fiber.Map{
		"tasks":    taskData,
	})
}

// fetchData handles database queries and decoding
func fetchData(ctx context.Context, coll *mongo.Collection, pipeline mongo.Pipeline, target interface{}, errChan chan error) {
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		errChan <- fmt.Errorf("error in fetching data: %v", err)
		return
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, target)
	if err != nil {
		errChan <- fmt.Errorf("error in decoding data: %v", err)
		return
	}

	errChan <- nil
}

/// Things to remember 
/// 1.1 -> first get all types from kt_m_types -> append to three columns
/// 1.2 -> prepare the overall json
/// 1.3 -> formulate roll down and roll up operation 