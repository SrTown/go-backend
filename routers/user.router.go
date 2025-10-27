package routers

import (
	"github.com/SrTown/go-backend/handlers"
	"github.com/SrTown/go-backend/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UserRouter(router fiber.Router, db *pgxpool.Pool) {
	userHandler := handlers.NewUserHandler(db)

	router.Get("/", userHandler.GetUsers)
	router.Get("/profile", userHandler.GetProfile)
	router.Post("/updatePassword", middlewares.ValidateUpdatePassword, userHandler.UpdatePassword)
	router.Get("/delete/:code", userHandler.DeleteUsers)
	router.Get("/:email", userHandler.GetOneUser)
}
