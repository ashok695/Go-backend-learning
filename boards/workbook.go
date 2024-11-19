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
	"reflect"
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
func createMapForRWT(taskTypeData []TaskTypeStrcut) map[primitive.ObjectID]TaskTypeStrcut{
	taskTypeMap:= make(map[primitive.ObjectID]TaskTypeStrcut)
	for _,taskType:= range taskTypeData {
		taskTypeMap[taskType.ID] = taskType
	}
	return taskTypeMap
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
func AssignTaskType(task *TaskStruct,taskTypeMap map[primitive.ObjectID]TaskTypeStrcut){
	fmt.Println("Hello")
	fmt.Println("type of ", reflect.TypeOf(task.TaskType))
		if taskSlice,ok := task.TaskType.(primitive.A);ok{
			if len(taskSlice) > 0 {
				if  tasktype,ok := taskSlice[0].(primitive.ObjectID);ok {
					fmt.Println("okkkkkkkk")
				if hi,exists := taskTypeMap[tasktype];exists{
					fmt.Println("is exists")
					task.TaskType = hi
				}
				}
			}
		}
}
// func AssignWorkStream(task *TaskStruct){
// 	if len(task.workstream) > 0 {

// 	}
// }
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
	appPort := app.Listen(":9000")
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
type TaskTypeStrcut struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}
func getData(c *fiber.Ctx)error{
	ctx:= context.Background()
	var subPhaseData []SubPhaseStruct
	var taskData []TaskStruct
	var userData []AssignedToStruct
	var statusData []StatusStruct
	var taskTypeData []TaskTypeStrcut

	subPhaseDBDetails:= client.Database("google-pt").Collection("kt_m_subphases")
	taskDBDetails:= client.Database("google-pt").Collection("kt_t_taskLists")
	userDBDetails:= client.Database("google-pt").Collection("kt_m_users")
	statusDBDetails:= client.Database("google-pt").Collection("kt_m_status")
	taskTypeDBDetails:= client.Database("google-pt").Collection("kt_m_types")

	subPhasePipeline := mongo.Pipeline{
		{{"$project",bson.D{{"_id",1},{"subPhaseName",1},{"__type",1},{"orderID",1}}}},
	}
	taskPipeline := mongo.Pipeline{
		{{"$match",bson.D{{"refBoardID",bson.D{{"$exists",false}}},{"skip",false}}}},
		{{"$project",bson.D{{"_id",1},{"title",1},{"status",1},{"assignedTo",1},{"orderID",1},{"plannedFrom",1},{"plannedTo",1},{"startedOn",1},{"completedOn",1},{"priority",1},{"role",1},{"workstream",1},{"type",1}}}},
	}
	userPipeline := mongo.Pipeline{
		{{"$project",bson.D{{"_id",1},{"fullName",1},{"email",1}}}},
	}
	statusPipeline := mongo.Pipeline{
		{{"$match",bson.D{{"workItem","Task"}}}},
		{{"$project",bson.D{{"_id",1},{"category",1},{"status",1}}}},
	}
	taskTypePipeline := mongo.Pipeline{
		{{"$match",bson.D{{"__type","tasktype"}}}},
		{{"$project",bson.D{{"_id",1},{"name",1}}}},
	}

	subPhaseChan := make(chan error)
	taskChan := make(chan error)
	userChan := make(chan error)
	statusChan := make(chan error)
	taskTypeChan := make(chan error)

	go func(){
		cursor,err := subPhaseDBDetails.Aggregate(ctx,subPhasePipeline)
		if err != nil {
			fmt.Println("Error in Getting Phase Data")
			subPhaseChan <- fmt.Errorf("error in getting Phase Data %v",err)
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx){
			var data SubPhaseStruct
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("error in decoding phase error data")
				subPhaseChan <- fmt.Errorf("Error in decoding subphase data %v",decodedError)
			}
			subPhaseData= append(subPhaseData,data)
		}
		subPhaseChan <- nil
	} ()
	go func(){
		cursor,err := taskDBDetails.Aggregate(ctx,taskPipeline)
		if err != nil {
			fmt.Println("Error in Getting Phase Data")
			taskChan <- fmt.Errorf("error in getting Phase Data %v",err)
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx){
			var data TaskStruct
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("error in decoding phase error data")
				taskChan <- fmt.Errorf("Error in decoding subphase data %v",decodedError)
			}
			if data.ActualStart == nil {
				data.ActualStart = ""
			}
			if data.ActualEnd == nil {
				data.ActualEnd = ""
			}
			taskData= append(taskData,data)
		}
		taskChan <- nil
	} ()
	go func(){
		cursor,err := userDBDetails.Aggregate(ctx,userPipeline)
		if err != nil {
			fmt.Println("Error in Getting Phase Data")
			userChan <- fmt.Errorf("error in getting Phase Data %v",err)
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx){
			var data AssignedToStruct
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("error in decoding phase error data")
				userChan <- fmt.Errorf("Error in decoding subphase data %v",decodedError)
			}
			userData= append(userData,data)
		}
		userChan <- nil
	} ()
	go func(){
		cursor,err := statusDBDetails.Aggregate(ctx,statusPipeline)
		if err != nil {
			fmt.Println("Error in Getting Phase Data")
			statusChan <- fmt.Errorf("error in getting Phase Data %v",err)
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx){
			var data StatusStruct
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("error in decoding phase error data")
				statusChan <- fmt.Errorf("Error in decoding subphase data %v",decodedError)
			}
			statusData= append(statusData,data)
		}
		statusChan <- nil
	} ()
	go func(){
		cursor,err := taskTypeDBDetails.Aggregate(ctx,taskTypePipeline)
		if err != nil {
			fmt.Println("Error in Getting Phase Data")
			taskTypeChan <- fmt.Errorf("error in getting Phase Data %v",err)
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx){
			var data TaskTypeStrcut
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("error in decoding phase error data")
				taskTypeChan <- fmt.Errorf("Error in decoding subphase data %v",decodedError)
			}
			taskTypeData= append(taskTypeData,data)
		}
		taskTypeChan <- nil
	} ()

	if err := <-taskChan; err != nil{
		fmt.Println("Error in suphase data")
	}
	if err := <-subPhaseChan; err != nil{
		fmt.Println("Error in suphase data")
	}
	if err := <-userChan; err != nil{
		fmt.Println("Error in suphase data")
	}
	if err := <-statusChan; err != nil{
		fmt.Println("Error in suphase data")
	}
	if err := <-taskTypeChan; err != nil{
		fmt.Println("Error in suphase data")
	}
	maps:=createMapForUserAndStatus(userData,statusData)
	mapsForRWT := createMapForRWT(taskTypeData)
	userMap:=maps.UserMap
	statusMap := maps.StatusMap
	for i := range taskData {
		task:=&taskData[i]
		assignOwnerAndStatus(task,userMap,statusMap)
		FindStartVariance(task)
		EndVariance(task)
		AssignTaskType(task,mapsForRWT)
	}
	return c.JSON(fiber.Map{
		"status":200,
		"msg":"Hello From Server",
		// "userMap":userMap,
		// "statusMap":statusMap,
		"data":taskData,

	})
}

/// Things to remember 
/// 1.1 -> first get all types from kt_m_types -> append to three columns
/// 1.2 -> prepare the overall json
/// 1.3 -> formulate roll down and roll up operation 