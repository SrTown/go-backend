package middlewares

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GetBearerToken(c *fiber.Ctx) error {
	// Get Authorization header
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"ok":            false,
			"token_expired": false,
			"message":       "Access denied. Bearer token missing.",
		})
	}

	parts := strings.Split(authHeader, " ")

	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"ok":            false,
			"token_expired": false,
			"message":       "Access denied. Bearer token missing.",
		})
	}

	token := parts[1]

	tokenKeyword := os.Getenv("muercielago-truora")
	if tokenKeyword == "" {
		tokenKeyword = "tokentest"
	}

	// Verificamos JWT token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusForbidden, "Invalid signing method")
		}
		return []byte(tokenKeyword), nil
	})

	if err != nil || !parsedToken.Valid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"ok":            false,
			"token_expired": true,
			"message":       "Access denied. The bearer token is invalid.",
		})
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if userID, exists := claims["id_user"]; exists {
			c.Locals("id_user", userID)
		}
		c.Locals("bearer_token", token)
	}

	return c.Next()
}

func ValidateRoutePrivate(c *fiber.Ctx) error {
	// Obtenemos el token
	token := c.Cookies("access_token")

	if token == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"ok":            false,
			"token_expired": false,
			"message":       "Access denied. Please sign in.",
		})
	}

	tokenKeyword := os.Getenv("muercielago-truora")
	if tokenKeyword == "" {
		tokenKeyword = "tokentest"
	}

	// Verificar JWT token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid signing method")
		}
		return []byte(tokenKeyword), nil
	})

	// Handle verification errors
	if err != nil || !parsedToken.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"ok":            false,
			"token_expired": true,
			"message":       "Access denied. Session token expired.",
		})
	}

	// Obtener datos del token
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		// Guardar ID en local (equivalente al req.id_user en Express)
		if userID, exists := claims["id_user"]; exists {
			c.Locals("id_user", userID)
		}
		if typeUser, exists := claims["type_user"]; exists {
			c.Locals("type_user", typeUser)
		}

		if c.Locals("id_user") == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"ok":            false,
				"token_expired": true,
				"message":       "Access denied. Session token expired.",
			})
		}
	}

	return c.Next()
}
