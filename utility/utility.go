package utility

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unicode"
)

func DeleteFile(filePath string) (bool, error) {
	err := os.Remove(filePath)
	if err == nil {
		return true, nil
	}
	if err.(*os.PathError).Err == syscall.ENOENT {
		return false, nil
	}
	return false, fmt.Errorf("could not delete the file: %w", err)
}

func GetFormattedErrorMessage(err error) string {
	stringBuilder := &strings.Builder{}
	messages := strings.Split(err.Error(), ": ")
	for i, message := range messages {
		runes := []rune(message)
		runes[0] = unicode.ToUpper(runes[0])
		message = string(runes)
		if i != 0 {
			stringBuilder.WriteString("> ")
		}
		stringBuilder.WriteString(message)
		if i == len(messages)-1 {
			stringBuilder.WriteString(".\n")
		} else {
			stringBuilder.WriteString(":\n")
		}
	}
	return stringBuilder.String()
}
