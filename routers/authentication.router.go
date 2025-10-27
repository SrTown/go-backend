package routers

import (
	"github.com/SrTown/go-backend/handlers"
	"github.com/SrTown/go-backend/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRouter(router fiber.Router, db *pgxpool.Pool) {
	authHandler := handlers.NewAuthHandler(db)

	router.Post("/login", authHandler.Login)
	router.Post("/signup", middlewares.ValidateSignup, authHandler.Signup)
	router.Post("/logout", authHandler.Logout)
	router.Post("/forgotPassword", authHandler.ForgotPassword)
}
