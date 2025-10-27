package handlers

import (
	"context"
	"fmt"

	"github.com/SrTown/go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApiHandler struct {
	DB *pgxpool.Pool
}

var allowedTables = map[string]bool{
	"users":                   true,
	"analyst_recommendations": true,
}

func NewApiHandler(db *pgxpool.Pool) *ApiHandler {
	return &ApiHandler{DB: db}
}

func (h *ApiHandler) GetData(c *fiber.Ctx) error {
	tableName := c.Params("tableName")

	// Validate if tha table is allowed
	if !allowedTables[tableName] {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": fmt.Sprintf("The table %s doesn't exist.", tableName),
		})
	}

	ctx := context.Background()

	// Parse query parameters
	queryParams := make(map[string]interface{})

	// Get all query parameters
	c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := string(key)
		valueStr := string(value)
		queryParams[keyStr] = valueStr
	})

	// Check if it's a count request
	isCountRequest := false
	if _, exists := queryParams["_count"]; exists {
		isCountRequest = true
		delete(queryParams, "_count")
	}

	// Call the query modifier
	qm := utils.ParseQueryModifier(queryParams)

	// Handle count request
	if isCountRequest {
		countSQL, countArgs, err := qm.BuildCountSQL(tableName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"ok":      false,
				"message": "Error building count query.",
			})
		}

		var count int
		err = h.DB.QueryRow(ctx, countSQL, countArgs...).Scan(&count)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"ok":      false,
				"message": "Contact the developer.",
				"error":   err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"ok":    true,
			"count": count,
		})
	}

	// Build SELECT query
	sqlQuery, args, err := qm.BuildSQL(tableName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": "Error building query.",
		})
	}

	// Execute query
	rows, err := h.DB.Query(ctx, sqlQuery, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": "Contact the developer.",
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	var records []map[string]interface{}

	// Scan all rows
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			continue
		}

		record := make(map[string]interface{})
		for i, value := range values {
			columnName := string(fieldDescriptions[i].Name)
			// Exclude password field
			if columnName != "password" {
				record[columnName] = value
			}
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"message": "Error reading data.",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"ok":    true,
		"count": len(records),
		"data":  records,
	})
}
