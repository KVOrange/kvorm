package kvorm

import (
	"strings"
	"unicode"
)

const SEPARATOR = "__"

func toSnakeCase(str string) string {
	var result string
	var words []string
	var lastPosition int
	runes := []rune(str)

	for i, character := range runes {
		if i == 0 {
			continue
		}

		if unicode.IsUpper(character) {
			words = append(words, str[lastPosition:i])
			lastPosition = i
		}
	}

	// Добавить последнее слово в slice
	words = append(words, str[lastPosition:])

	for _, word := range words {
		if result != "" {
			result += "_"
		}
		result += strings.ToLower(word)
	}

	return result
}

func nestedMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		keys := strings.Split(key, ".")
		insertRecursively(result, keys, value)
	}

	return result
}

func nestedMapSlice(slice []map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	for _, item := range slice {
		nestedItem := nestedMap(item)
		result = append(result, nestedItem)
	}

	return result
}

func insertRecursively(data map[string]interface{}, keys []string, value interface{}) {
	if len(keys) == 1 {
		data[keys[0]] = value
		return
	}

	if _, exists := data[keys[0]]; !exists {
		data[keys[0]] = make(map[string]interface{})
	}

	subMap, _ := data[keys[0]].(map[string]interface{})
	insertRecursively(subMap, keys[1:], value)
}
