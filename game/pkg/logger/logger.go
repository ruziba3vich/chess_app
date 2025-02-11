package logger

import (
	"log"
	"os"
)

// NewLogger creates a logger that writes to both a file and the console
func NewLogger(logFilePath string) (*log.Logger, error) {
	// Open the log file in append mode, create if it doesn't exist
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Create a multi-writer to write logs to both the file and stdout
	multiWriter := log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)

	return multiWriter, nil
}
