package utils

import (
	"log"
	"os"
)

func HomeDirOrEmpty(suffix string) string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Printf("WARN: can not get home dir path: %s", err)
		return ""
	}
	return dirname + suffix
}
