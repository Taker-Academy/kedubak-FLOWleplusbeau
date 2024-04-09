package router

import (
	"containerized-go-app/hash"
	"containerized-go-app/jwt"
	"containerized-go-app/models"
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

func UserRoutes(app *fiber.App, db *mongo.Database) {
	user := app.Group("/user", func(c *fiber.Ctx) error {
		return c.Next()
	})
	GetUser(db, user)
	EditUser(db, user)
	DeleteUser(db, user)
}

func GetUser(db *mongo.Database, user fiber.Router) {
	user.Get("/me", func(c *fiber.Ctx) error {
		userID, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "Unauthorized",
			})
		}

		userCollection := db.Collection("User")

		objId, _ := primitive.ObjectIDFromHex(userID)
		user := models.User{}
		err = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "Unauthorized",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"email":     user.Email,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
			},
		})
	})
}

func EditUser(db *mongo.Database, user fiber.Router) {
	user.Put("/edit", func(c *fiber.Ctx) error {
		userID, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "Unauthorized",
			})
		}

		userCollection := db.Collection("User")

		objId, _ := primitive.ObjectIDFromHex(userID)
		user := models.User{}
		err = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "Unauthorized",
			})
		}

		var userUpdate models.User
		if err := c.BodyParser(&userUpdate); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"ok":    false,
				"error": "Bad Request",
			})
		}

		update := bson.M{}
		if userUpdate.FirstName != "" {
			update["firstName"] = userUpdate.FirstName
		}
		if userUpdate.LastName != "" {
			update["lastName"] = userUpdate.LastName
		}
		if userUpdate.Email != "" {
			update["email"] = userUpdate.Email
		}
		if userUpdate.Password != "" {
			update["password"], err = hash.HashPassword(userUpdate.Password)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"ok":    false,
					"error": "Internal Server Error",
				})
			}
		}

		_, err = userCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": update})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"email":     userUpdate.Email,
				"firstName": userUpdate.FirstName,
				"lastName":  userUpdate.LastName,
			},
		})
	})
}

func DeleteUser(db *mongo.Database, user fiber.Router) {
	user.Delete("/delete", func(c *fiber.Ctx) error {
		userID, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "Unauthorized",
			})
		}

		userCollection := db.Collection("User")

		objId, _ := primitive.ObjectIDFromHex(userID)
		user := models.User{}
		err = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"ok":    false,
				"error": "Not Found",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"email":     user.Email,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"removed":   true,
			},
		})
	})
}
