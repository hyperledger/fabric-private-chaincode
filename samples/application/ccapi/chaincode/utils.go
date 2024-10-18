package chaincode

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

func extractStatusCode(msg string) int {
	re := regexp.MustCompile(`Code:\s*\((\d+)\)`)

	matches := re.FindStringSubmatch(msg)
	if len(matches) == 0 {
		fmt.Println("No status code found in message")
		return http.StatusInternalServerError
	}

	statusCode, err := strconv.Atoi(matches[1])
	if err != nil {
		fmt.Println("Failed to parse string to int when extracting status code")
		return http.StatusInternalServerError
	}

	return statusCode
}
