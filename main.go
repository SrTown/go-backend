package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/SrTown/go-backend/middlewares"
	"github.com/SrTown/go-backend/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Todo struct {
	ID        int    `json:"id"`
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

func main() {
	fmt.Println("Holaa")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	COCKROCH_URI := "postgresql://carlos:43-hcmxO4FtfI1ZAZj3WDA@aging-avocet-9538.jxf.gcp-us-east1.cockroachlabs.cloud:26257/stock-info?sslmode=verify-full"
	PORT := os.Getenv("PORT")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, COCKROCH_URI)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected successfully to CockroachDB")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Authorization",
		ExposeHeaders:    "Content-Length, Authorization",
		AllowCredentials: false,
	}))

	//Initiall routes declaration with middlewares
	userRoutes := app.Group("/user", middlewares.ValidateRoutePrivate, middlewares.GetBearerToken)
	authRoutes := app.Group("/auth")
	apiRoutes := app.Group("/api", middlewares.ValidateRoutePrivate, middlewares.GetBearerToken)
	// apipubRoutes:= app.Group("/apipub")

	//Creation of sub-routes
	routers.UserRouter(userRoutes, pool)
	routers.AuthRouter(authRoutes, pool)
	routers.ApiRouter(apiRoutes, pool)
	// routers.ApipubRouter(apipubRoutes)

	log.Fatal(app.Listen(":" + PORT))
}
