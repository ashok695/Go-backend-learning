package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/contrib/websocket"
)

func main(){
	app:= fiber.New()
	app.Get("/",getData)
	portError := app.Listen(":9000")
	if portError!=nil {
		fmt.Println("Error in POrt Opening")
	}
}

func getData(c *fiber.Ctx)error{
	return c.JSON(fiber.Map{"status":200,"msg":"Hello from getdata server"})
}