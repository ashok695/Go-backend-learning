package main

import(
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
	app := fiber.New()
	MONGO_URL := "mongodb+srv://qas-app-rwuser:fSXQNZ7G9B38n4xi@quality-app.ss86w.mongodb.net/"
	clientOptions := client.options.ApplyURI()
}