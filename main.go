package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	// "os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	// Create a new engine
	engine := html.New("./view", ".html")

	// Fiber instance
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// setup s3 uploader
	client := s3.NewFromConfig(cfg)

	uploader := manager.NewUploader(client)

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	app.Post("/", func(c *fiber.Ctx) error {
		// Get first file from form field "document":
		file, err := c.FormFile("upload")
		if err != nil {
			return err
		}
		// open the file
		f, err := file.Open()

		if err != nil {
			return err
		}

		// result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		// 	Bucket: aws.String("grain"),
		// 	Key:    aws.String("grains/gn" + file.Filename),
		// 	Body:   f,
		// 	ACL:    "public-read",
		// })
		// fmt.Println(os.Getenv("AWS_BUCKET"))
		result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String("grain"),
			Key:    aws.String(file.Filename),
			Body:   f,
			ACL:    "public-read",
		}, func(u *manager.Uploader) {
			u.PartSize = 6 * 1024 * 1024 // Override the PartSize to 6 MiB
		})

		if err != nil {
			var mu manager.MultiUploadFailure
			if errors.As(err, &mu) {
				// Process error and its associated UploadID
				fmt.Println("Error:", mu)
				uploadID := mu.UploadID() // Retrieve the associated UploadID
				// You can use uploadID for further actions, like resuming the upload.
				return c.JSON(fiber.Map{
					"status":  "error",
					"message": "File uploaded failed",
					"file":    uploadID,
				})
			} else {
				// Process error generically
				fmt.Println("Error:", err.Error())
			}
		}

		fmt.Println(result, err)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "File uploaded successfully",
			"url":     result.Location,
		})

		// Save file to root directory:
		// c.SaveFile(file, "public/uploads/p"+file.Filename)
		// return c.Render("index", fiber.Map{})
	})

	// Create 'uploads' directory if not exists
	// if _, err := os.Stat("public/uploads"); os.IsNotExist(err) {
	// 	os.Mkdir("public/uploads", 0755)
	// }

	// Static files
	app.Static("/", "./public")

	// Start server
	app.Listen(":3000")
}
