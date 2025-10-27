package handlers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	DB *pgxpool.Pool
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	UserType string `json:"user_type"`
	Status   bool   `json:"status"`
}

func NewUserHandler(db *pgxpool.Pool) *UserHandler {
	return &UserHandler{DB: db}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get all users"})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("id_user")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"ok":      false,
			"message": "User ID not found in context.",
		})
	}

	ctx := context.Background()

	query := `
		SELECT id, email, name, user_type, status
		FROM users
		WHERE id = $1
	`

	var user User
	err := h.DB.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.UserType,
		&user.Status,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"ok":      false,
				"message": "Unable to get user information.",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": fmt.Sprintf("Database error: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"ok":   true,
		"data": user,
	})
}

func (h *UserHandler) GetOneUser(c *fiber.Ctx) error {
	email := c.Params("email")
	return c.JSON(fiber.Map{"email": email})
}

func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Password updated"})
}

func (h *UserHandler) DeleteUsers(c *fiber.Ctx) error {
	code := c.Params("code")
	return c.JSON(fiber.Map{"message": "User deleted", "code": code})
}
