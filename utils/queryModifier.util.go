package utils

import (
	"strconv"
	"strings"
)

type QueryModifier struct {
	Query            map[string]interface{}
	Offset           *int
	Limit            *int
	Attributes       []string
	StrictAttributes []string
	IncludeQuery     map[string]map[string]interface{}
	OrderBy          []OrderByClause
	Distinct         interface{}
	LikeConditions   map[string]string
	InConditions     map[string][]interface{}
	NullOrConditions map[string][]interface{}
}

type OrderByClause struct {
	Field     string
	Direction string // ASC or DESC
}

func NewQueryModifier(queryParams map[string]interface{}) *QueryModifier {
	return &QueryModifier{
		Query:            make(map[string]interface{}),
		IncludeQuery:     make(map[string]map[string]interface{}),
		OrderBy:          []OrderByClause{},
		LikeConditions:   make(map[string]string),
		InConditions:     make(map[string][]interface{}),
		NullOrConditions: make(map[string][]interface{}),
	}
}

func ParseQueryModifier(query map[string]interface{}) *QueryModifier {
	qm := NewQueryModifier(query)

	for key, value := range query {
		// Handle JOIN queries (e.g., "table->column")
		if strings.Contains(key, "->") {
			parts := strings.Split(key, "->")
			join := parts[0]
			attribute := parts[1]

			if qm.IncludeQuery[join] == nil {
				qm.IncludeQuery[join] = make(map[string]interface{})
			}
			qm.IncludeQuery[join][attribute] = value
			continue
		}

		if strValue, ok := value.(string); ok {
			// LIKEE
			if strings.HasPrefix(strValue, "_lk") && strings.HasSuffix(strValue, "_lk") {
				trimmedValue := strings.TrimPrefix(strValue, "_lk")
				trimmedValue = strings.TrimSuffix(trimmedValue, "_lk")
				qm.LikeConditions[key] = trimmedValue
				continue
			}

			// Comas para ser un IN o OR with NULLs
			if strings.Contains(strValue, ",") {
				values := strings.Split(strValue, ",")
				var cleanValues []interface{}
				hasNull := false

				for _, v := range values {
					trimmed := strings.TrimSpace(v)
					if trimmed == "_null" {
						hasNull = true
					} else {
						if num, err := strconv.ParseFloat(trimmed, 64); err == nil {
							cleanValues = append(cleanValues, num)
						} else {
							cleanValues = append(cleanValues, trimmed)
						}
					}
				}

				if hasNull {
					// OR
					qm.NullOrConditions[key] = cleanValues
				} else {
					// IN
					qm.InConditions[key] = cleanValues
				}
				continue
			}
		}

		// Keywords
		switch key {
		case "_offset":
			if offset := parseIntValue(value); offset != nil {
				qm.Offset = offset
			}
		case "_limit":
			if limit := parseIntValue(value); limit != nil {
				qm.Limit = limit
			}
		case "_cmp":
			qm.Attributes = parseStringArrayValue(value, "password")
		case "_scmp":
			qm.StrictAttributes = parseStringArrayValue(value, "password")
		case "_orderby":
			// Will be combined with _ordertype below
		case "_ordertype":
			// Will be combined with _orderby below
		case "_distinct":
			qm.Distinct = value
		case "_cache":
			// Ignore cache parameter
		default:
			// Add to regular query
			qm.Query[key] = value
		}
	}

	// Handle ORDER BY (needs both _orderby and _ordertype)
	if orderBy, hasOrderBy := query["_orderby"]; hasOrderBy {
		if orderType, hasOrderType := query["_ordertype"]; hasOrderType {
			if orderByStr, ok := orderBy.(string); ok {
				if orderTypeStr, ok := orderType.(string); ok {
					qm.OrderBy = append(qm.OrderBy, OrderByClause{
						Field:     orderByStr,
						Direction: strings.ToUpper(orderTypeStr),
					})
				}
			}
		}
	}

	return qm
}

func parseIntValue(value interface{}) *int {
	switch v := value.(type) {
	case string:
		if num, err := strconv.Atoi(v); err == nil {
			return &num
		}
	case []interface{}:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				if num, err := strconv.Atoi(str); err == nil {
					return &num
				}
			}
		}
	case int:
		return &v
	case float64:
		num := int(v)
		return &num
	}
	return nil
}

func parseStringArrayValue(value interface{}, exclude string) []string {
	var result []string

	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok && str != exclude {
				result = append(result, str)
			}
		}
	case string:
		if v != exclude {
			result = append(result, v)
		}
	}

	return result
}

// BuildSQL generates SQL query from the modifier
func (qm *QueryModifier) BuildSQL(tableName string) (string, []interface{}, error) {
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	// Handle regular query conditions
	for key, value := range qm.Query {
		whereClauses = append(whereClauses, key+" = $"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	// Handle LIKE conditions
	for key, value := range qm.LikeConditions {
		whereClauses = append(whereClauses, key+" LIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+value+"%")
		argIndex++
	}

	// Handle IN conditions
	for key, values := range qm.InConditions {
		placeholders := []string{}
		for _, v := range values {
			placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
			args = append(args, v)
			argIndex++
		}
		whereClauses = append(whereClauses, key+" IN ("+strings.Join(placeholders, ", ")+")")
	}

	// Handle NULL OR conditions
	for key, values := range qm.NullOrConditions {
		placeholders := []string{}
		for _, v := range values {
			placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
			args = append(args, v)
			argIndex++
		}
		condition := "(" + key + " IN (" + strings.Join(placeholders, ", ") + ") OR " + key + " IS NULL)"
		whereClauses = append(whereClauses, condition)
	}

	// Build SELECT clause
	selectClause := "*"
	if len(qm.StrictAttributes) > 0 {
		selectClause = strings.Join(qm.StrictAttributes, ", ")
	} else if len(qm.Attributes) > 0 {
		selectClause = strings.Join(qm.Attributes, ", ")
	}

	// Build DISTINCT
	distinctClause := ""
	if qm.Distinct != nil {
		distinctClause = "DISTINCT "
	}

	// Build base query
	sqlQuery := "SELECT " + distinctClause + selectClause + " FROM " + tableName

	// Add WHERE clause
	if len(whereClauses) > 0 {
		sqlQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Add ORDER BY
	if len(qm.OrderBy) > 0 {
		orderClauses := []string{}
		for _, order := range qm.OrderBy {
			orderClauses = append(orderClauses, order.Field+" "+order.Direction)
		}
		sqlQuery += " ORDER BY " + strings.Join(orderClauses, ", ")
	}

	// Add LIMIT
	if qm.Limit != nil {
		sqlQuery += " LIMIT $" + strconv.Itoa(argIndex)
		args = append(args, *qm.Limit)
		argIndex++
	}

	// Add OFFSET
	if qm.Offset != nil {
		sqlQuery += " OFFSET $" + strconv.Itoa(argIndex)
		args = append(args, *qm.Offset)
		argIndex++
	}

	return sqlQuery, args, nil
}

func (qm *QueryModifier) BuildCountSQL(tableName string) (string, []interface{}, error) {
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	// Handle regular query conditions
	for key, value := range qm.Query {
		whereClauses = append(whereClauses, key+" = $"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	// Handle LIKE conditions
	for key, value := range qm.LikeConditions {
		whereClauses = append(whereClauses, key+" LIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+value+"%")
		argIndex++
	}

	// Handle IN conditions
	for key, values := range qm.InConditions {
		placeholders := []string{}
		for _, v := range values {
			placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
			args = append(args, v)
			argIndex++
		}
		whereClauses = append(whereClauses, key+" IN ("+strings.Join(placeholders, ", ")+")")
	}

	// Handle NULL OR conditions
	for key, values := range qm.NullOrConditions {
		placeholders := []string{}
		for _, v := range values {
			placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
			args = append(args, v)
			argIndex++
		}
		condition := "(" + key + " IN (" + strings.Join(placeholders, ", ") + ") OR " + key + " IS NULL)"
		whereClauses = append(whereClauses, condition)
	}

	// Build COUNT query
	countClause := "COUNT(*)"
	if qm.Distinct != nil {
		if distinctStr, ok := qm.Distinct.(string); ok {
			countClause = "COUNT(DISTINCT " + distinctStr + ")"
		}
	}

	sqlQuery := "SELECT " + countClause + " FROM " + tableName

	// Add WHERE clause
	if len(whereClauses) > 0 {
		sqlQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return sqlQuery, args, nil
}
