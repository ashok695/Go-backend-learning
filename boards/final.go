package main

import(
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"context"
	// "sync"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
var (
	client *mongo.Client
	// wg sync.WaitGroup
)
func main(){
	ctx:= context.Background()
	var err error
	app:= fiber.New()
	MONGODB_URL :="mongodb+srv://qas-app-rwuser:fSXQNZ7G9B38n4xi@quality-app.ss86w.mongodb.net/"
	options:=options.Client().ApplyURI(MONGODB_URL)
	client,err = mongo.Connect(ctx,options)
	if err != nil {
		fmt.Println("DB CONNECTION SUCCESS")
	}
	pingErr := client.Ping(ctx,nil)
	if pingErr != nil {
		fmt.Println("ERROR IN PINGING DB")
	}
	defer client.Disconnect(context.Background())
	app.Get("/",getData)
	portError := app.Listen(":8000")
	if portError != nil {
		fmt.Println("Error in Connecting Port")
	}
}
// func assignStatus(task *TaskDetailsStruct, statusData []StatusDetailStruct){
// 	for _,status := range statusData {
// 		if task.Status == status.Id {
// 			task.Status = status
// 		}
// 	}
// }
// func assignOwner(task *TaskDetailsStruct, userData []AssignedToDetails){
// 	for _,user := range userData {
// 		if task.AssignedTo == user.Id {
// 			task.AssignedTo = user
// 		}
// 	}
// }
func createAssignedToMap(userData []AssignedToDetails)map[primitive.ObjectID]AssignedToDetails{
	assignedToMap := make(map[primitive.ObjectID]AssignedToDetails)
	for _,assigned := range userData {
		assignedToMap[assigned.Id] = assigned
	}
	return assignedToMap
}
func createStatusMap(statusData []StatusDetailStruct)map[primitive.ObjectID]StatusDetailStruct{
	statusMap := make(map[primitive.ObjectID]StatusDetailStruct)
		for _,status := range statusData {
			statusMap[status.Id] = status
	}
	return statusMap
}
func assignedOwner(task *TaskDetailsStruct, userMap map[primitive.ObjectID]AssignedToDetails) {
    if assignedToID, ok := task.AssignedTo.(primitive.ObjectID); ok {
        if user, exists := userMap[assignedToID]; exists {
            task.AssignedTo = user
        }
    }
}
func assignStatus(task *TaskDetailsStruct,statusMap map[primitive.ObjectID]StatusDetailStruct){
	if statusOK, ok := task.Status.(primitive.ObjectID); ok {
		if status,exists := statusMap[statusOK];exists {
			task.Status = status
		}
	}
}
func assignTasks(taskData []TaskDetailsStruct, boardData []BoardDetailsStruct) {
    boardMap := make(map[primitive.ObjectID]*BoardDetailsStruct)
    for i := range boardData {
        boardMap[boardData[i].Id] = &boardData[i]
    }

    for _, task := range taskData {
        if board, exists := boardMap[task.RefBoardID[0]]; exists {
            board.Tasks = append(board.Tasks, task)
        }
    }
}

type BoardDetailsStruct struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title string `json:"title" bson:"title"`
	Category string `json:"category" bson:"category"`
	Tasks []TaskDetailsStruct `json:"tasks" bson:"tasks"`
}
type TaskDetailsStruct struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title string `json:"title" bson:"title"`
	Status interface{} `json:"status" bson:"status"`
	AssignedTo interface{} `json:"assignedTo" bson:"assignedTo"`
	RefBoardID []primitive.ObjectID `json:"refBoardID" bson:"refBoardID"`
}
type StatusDetailStruct struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Category string `json:"category" bson:"category"`
	Status string `json:"status" bson:"status"`
}
type AssignedToDetails struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FullName string `json:"fullName" bson:"fullName"`
	Email string `json:"email" bson:"email"`
}
func getData(c *fiber.Ctx)error{
	// wg.Add(4)
	ctx:=context.Background()
	var boardData []BoardDetailsStruct
	var taskData []TaskDetailsStruct
	var statusData []StatusDetailStruct
	var userData []AssignedToDetails

	boardDBDetails := client.Database("kternai-tp0").Collection("kt_m_boards")
	taskDBDetails := client.Database("kternai-tp0").Collection("kt_t_taskLists")
	statusDbDetails := client.Database("kternai-tp0").Collection("kt_m_status")
	userDBDetails := client.Database("kternai-tp0").Collection("kt_m_users")

	boardDataPipeline := mongo.Pipeline{
		{{"$match",bson.D{{"category","custom"},{"active",true},}}},
		{{"$project", bson.D{{"_id",1},{"title",1},{"category",1},}}},
	}
	taskDataPipeline := mongo.Pipeline{
		{{"$match",bson.D{{"refBoardID",bson.D{{"$exists",true}}},{"skip",false},{"board","custom"}}}},
		{{"$project", bson.D{{"_id",1},{"title",1},{"status",1},{"assignedTo",1},{"refBoardID",1},}}},
	}
	statusDataPipeline := mongo.Pipeline{
		{{"$match",bson.D{{"workItem","Task"}}}},
		{{"$project",bson.D{{"_id",1},{"category",1},{"status",1},{"color",1}}}},
	}
	userDataPipeline := mongo.Pipeline{
		{{"$project",bson.D{{"_id",1},{"fullName",1},{"email",1}}}},
	}

	boardChan := make(chan error)
	taskChan := make(chan error)
	statusChan:=make(chan error)
	userChan:=make(chan error)

	go func(){
		cursor,err := boardDBDetails.Aggregate(ctx,boardDataPipeline)
		defer cursor.Close(context.Background())
		if err != nil{
			fmt.Println("Error in getting Data")
			boardChan <- fmt.Errorf("Error in decoding board Data %v",err)
		}
		for cursor.Next(ctx){
			var data BoardDetailsStruct
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("error in decoding data")
				boardChan <- fmt.Errorf("Error in decoding board Data %v",decodedError)
			}
			// fmt.Println("data",data)
			boardData = append(boardData,data)
		}
		boardChan <- nil
		// defer wg.Done()
	}()
	go func(){
		cursor,err := taskDBDetails.Aggregate(ctx,taskDataPipeline)
		defer cursor.Close(context.Background())
		if err != nil{
			fmt.Println("Error in getting Data")
			taskChan <- fmt.Errorf("Error in decoding board Data %v",err)
		}
		for cursor.Next(ctx){
			var data TaskDetailsStruct
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("error in decoding data")
				taskChan <- fmt.Errorf("Error in decoding board Data %v",decodedError)
			}
			taskData = append(taskData,data)
		}
		taskChan <- nil
		// defer wg.Done()
	}()
	go func(){
		cursor,err := statusDbDetails.Aggregate(ctx,statusDataPipeline)
		if err != nil {
			fmt.Println("Error in Getting DB details")
			statusChan <- fmt.Errorf("Error in getting Db details %v",err)
		}
		for cursor.Next(ctx){
		var data StatusDetailStruct
		decodedError := cursor.Decode(&data)
		if decodedError != nil {
			fmt.Println("Error in Decoding data")
			statusChan <- fmt.Errorf("Error in Decoding Data %v",decodedError)
			}
		statusData = append(statusData,data)
		}
		statusChan <- nil
		// defer wg.Done()
	}()
	go func (){
		cursor,err := userDBDetails.Aggregate(ctx,userDataPipeline)
		if err != nil {
			fmt.Println("Error in user Data")
			userChan <- fmt.Errorf("error in getting user data %v",err)
		}
		for cursor.Next(ctx){
			var data AssignedToDetails
			decodedError := cursor.Decode(&data)
			if decodedError != nil {
				fmt.Println("Error in decoding user data")
				userChan <- fmt.Errorf("error in getting user data %v",err)
			}
			userData = append(userData,data)
		}
		userChan <- nil
		// defer wg.Done()
	}()

	if err := <- boardChan ; err!= nil {
		fmt.Println("error in channel",err)
	}
	if err := <- taskChan ; err!= nil {
		fmt.Println("error in channel",err)
	}
	if err := <- statusChan; err!= nil {
		fmt.Println("Error in status chan")
	}
	if err := <- userChan; err!= nil {
		fmt.Println("Error in userChan",err)
	}
	userMap := createAssignedToMap(userData)
	statusMap := createStatusMap(statusData)
	for i := range taskData {
		assignedOwner(&taskData[i],userMap)
		assignStatus(&taskData[i],statusMap)
	}
	assignTasks(taskData,boardData)
	// fmt.Println("Taskmap is ",userMap)
	// for i := range taskData{
	// 	assignStatus(&taskData[i],statusData)
	// 	assignOwner(&taskData[i],userData)
	// }
	// for i,board := range boardData {
	// 	for _,task := range taskData {
	// 		if task.RefBoardID[0] == board.Id {
	// 			boardData[i].Tasks=append(boardData[i].Tasks,task)
	// 		}
	// 	}
	// }
	// wg.Wait()
	return c.JSON(fiber.Map{
		"status":200,
		"boardData":boardData,
		// "taskData":taskData,
		// "statusData":statusData,
		// "userData":userData,
		// "userMap":userMap,
	})
}