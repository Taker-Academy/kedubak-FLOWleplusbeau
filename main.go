package main

import (
	"containerized-go-app/router"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

func connectToDB() (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil, errors.New("MONGO_URI is not set")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return nil, errors.New("DB_NAME is not set")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	return db, err
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	// load env variables
	err := godotenv.Load()
	if err != nil {
		return err
	}

	//init database
	db, err := connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	//defer disconnect
	defer db.Client().Disconnect(context.Background())

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	router.AuthRoutes(app, db)
	router.UserRoutes(app, db)
	router.PostRoutes(app, db)

	app.Listen(":8080")

	return nil
}
