package strings

import (
	"regexp"
	"strings"
)

var reservedWords = map[string]bool{
	"SELECT": true,
	"TABLE":  true,
	"INSERT": true,
	"DELETE": true,
	"UPDATE": true,
	"FROM":   true,
	"WHERE":  true,
	"GROUP":  true,
	"HAVING": true,
	"ORDER":  true,
	"BY":     true,
	"LIMIT":  true,
	"OFFSET": true,
}

func IsValidTableName(tableName string) bool {
	if len(tableName) == 0 || len(tableName) > 20 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, tableName)
	if !match {
		return false
	}
	if reservedWords[strings.ToUpper(tableName)] {
		return false
	}
	return true
}
