package middlewares

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	UserType string `json:"user_type" validate:"required"`
}

type UpdatePasswordRequest struct {
	Password    string `json:"password" validate:"required,min=6"`
	NewPassword string `json:"newPassword" validate:"required,min=6"`
}

var passwordErrorMessages = map[string]string{
	"Password.required":    "The password is mandatory.",
	"Password.min":         "The password must contain at least 6 characters.",
	"NewPassword.required": "The new password is mandatory.",
	"NewPassword.min":      "The new password must contain at least 6 characters.",
}

func ValidateUpdatePassword(c *fiber.Ctx) error {
	var body UpdatePasswordRequest

	// Boddy parser
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":     false,
			"errors": []string{"Invalid request body."},
		})
	}

	// Instanciamos validator
	validate := validator.New()
	err := validate.Struct(body)

	if err != nil {
		var errors []string

		for _, err := range err.(validator.ValidationErrors) {
			fieldName := err.Field()
			tag := err.Tag()
			key := fieldName + "." + tag

			// Seteo de mensajes
			if msg, exists := passwordErrorMessages[key]; exists {
				errors = append(errors, msg)
			} else {
				errors = append(errors, err.Error())
			}
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":     false,
			"errors": errors,
		})
	}

	c.Locals("updatePasswordData", body)

	return c.Next()
}

func ValidateSignup(c *fiber.Ctx) error {
	var body SignupRequest

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"message": "Invalid request body.",
		})
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		errors := make([]string, 0)

		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Email":
				if err.Tag() == "required" {
					errors = append(errors, "Email is required.")
				} else if err.Tag() == "email" {
					errors = append(errors, "Invalid email format.")
				}
			case "Name":
				errors = append(errors, "Name is required.")
			case "Password":
				if err.Tag() == "required" {
					errors = append(errors, "Password is required.")
				} else if err.Tag() == "min" {
					errors = append(errors, "Password must be at least 6 characters.")
				}
			case "UserType":
				errors = append(errors, "User type is required.")
			}
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":     false,
			"errors": errors,
		})
	}

	c.Locals("signupData", body)
	return c.Next()
}
