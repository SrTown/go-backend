package routers

import (
	"github.com/SrTown/go-backend/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ApiRouter(router fiber.Router, db *pgxpool.Pool) {
	apiHandler := handlers.NewApiHandler(db)

	router.Get("/:tableName", apiHandler.GetData)
}
