package handlers

import (
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB *pgxpool.Pool
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	UserType string `json:"user_type" validate:"required"`
}

type UserLogin struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"-"` // Para que no salga en el json
	UserType string `json:"user_type"`
	Status   bool   `json:"status"`
}

type ForgotPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	NewPassword string `json:"newPassword" validate:"required,min=6"`
}

func NewAuthHandler(db *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{DB: db}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var loginData LoginRequest

	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"message": "Invalid request body.",
		})
	}

	ctx := context.Background()

	// Search by email
	query := `
		SELECT id, email, name, password, user_type, status
		FROM users
		WHERE email = $1
	`

	var user UserLogin
	err := h.DB.QueryRow(ctx, query, loginData.Email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.UserType,
		&user.Status,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"ok":      false,
				"message": "Invalid credentials",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"ok":      false,
			"message": "Unable to sign in. DB error.",
			"error":   err.Error(),
		})
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"ok":      false,
			"message": "Invalid credentials",
		})
	}

	// Check if user is active
	if !user.Status {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"ok":      false,
			"message": "Access denied. User deleted.",
		})
	}

	// Create JWT token
	tokenKeyword := os.Getenv("muercielago-truora")
	if tokenKeyword == "" {
		tokenKeyword = "tokentest"
	}

	claims := jwt.MapClaims{
		"id_user":   user.ID,
		"type_user": user.UserType,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenKeyword))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": "Failed to create token",
		})
	}

	// Set cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "User logged successfully.",
		"token":   tokenString,
	})
}

func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var signupData SignupRequest

	if err := c.BodyParser(&signupData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"message": "Invalid request body.",
		})
	}

	ctx := context.Background()

	// Check if user already exists
	var existingEmail string
	checkQuery := `SELECT email FROM users WHERE email = $1`
	err := h.DB.QueryRow(ctx, checkQuery, signupData.Email).Scan(&existingEmail)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"ok":      false,
			"message": "The user already exists.",
		})
	}

	if err != pgx.ErrNoRows {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"ok":      false,
			"message": "Unable to register new user. DB error.",
			"error":   err.Error(),
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signupData.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": "Failed to encrypt password.",
		})
	}

	insertQuery := `
		INSERT INTO users (email, name, password, user_type, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var newUserID string
	err = h.DB.QueryRow(
		ctx,
		insertQuery,
		signupData.Email,
		signupData.Name,
		string(hashedPassword),
		signupData.UserType,
		true,
	).Scan(&newUserID)

	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"ok":      false,
			"message": "Unable to register new user. DB error.",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "User registered successfully.",
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear the access_token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Set to past time to delete
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "Session closed successfully.",
	})
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var forgotData ForgotPasswordRequest

	if err := c.BodyParser(&forgotData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"message": "Invalid request body.",
		})
	}

	ctx := context.Background()

	var userID string
	checkQuery := `SELECT id FROM users WHERE email = $1`
	err := h.DB.QueryRow(ctx, checkQuery, forgotData.Email).Scan(&userID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"ok":      false,
				"message": "There is not a user with that email address.",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": "Database error.",
			"error":   err.Error(),
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(forgotData.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": "Failed to encrypt password.",
		})
	}

	updateQuery := `
		UPDATE users 
		SET password = $1, updated_at = current_timestamp()
		WHERE email = $2
	`

	_, err = h.DB.Exec(ctx, updateQuery, string(hashedPassword), forgotData.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": "Failed to update password.",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "Password updated successfully.",
	})
}
