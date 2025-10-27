package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/SrTown/go-backend/middlewares"
	"github.com/SrTown/go-backend/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	app  *fiber.App
	pool *pgxpool.Pool
	once sync.Once
)

// Initialize app once (connection pooling for serverless)
func init() {
	once.Do(func() {
		cockroachURI := os.Getenv("COCKROACH_URI")
		if cockroachURI == "" {
			log.Fatal("COCKROACH_URI environment variable is not set")
		}

		config, err := pgxpool.ParseConfig(cockroachURI)
		if err != nil {
			log.Fatal("Failed to parse database config:", err)
		}

		config.MaxConns = 5 // Low number for serverless
		config.MinConns = 1
		config.MaxConnLifetime = 300 // 5 minutes
		config.MaxConnIdleTime = 60  // 1 minute

		ctx := context.Background()
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		// Create connection pool
		// ctx := context.Background()
		// var err error
		// pool, err = pgxpool.New(ctx, cockroachURI)
		// if err != nil {
		// 	log.Fatal("Failed to connect to database:", err)
		// }

		// Test connection
		err = pool.Ping(ctx)
		if err != nil {
			log.Fatal("Failed to ping database:", err)
		}

		log.Println("Connected successfully to CockroachDB")

		// Create Fiber app
		app = fiber.New(fiber.Config{
			ServerHeader: "Vercel",
		})

		// Setup routes
		userRoutes := app.Group("/user", middlewares.ValidateRoutePrivate)
		routers.UserRouter(userRoutes, pool)

		authRoutes := app.Group("/auth")
		routers.AuthRouter(authRoutes, pool)

		apiRoutes := app.Group("/api", middlewares.GetBearerToken)
		routers.ApiRouter(apiRoutes, pool)
	})
}

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	adaptor.FiberApp(app)(w, r)
}
