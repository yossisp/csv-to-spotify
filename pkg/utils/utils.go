package utils

import (
	"fmt"
	"log"
	"time"
)

type LogCb func(format string, v ...interface{})

// SleepFor sleep for n seconds
func SleepFor(seconds int) {
	time.Sleep(time.Duration(time.Duration(seconds) * time.Second))
}

// LogPanic logs panic error without exiting
func LogPanic(message string) {
	log.Println("LogPanic")
	if err := recover(); err != nil {
		log.Println("Recovered in ", message, err)
	}
}

// NewLogger logs within a file
func NewLogger(logPrefix string) LogCb {
	return func(format string, v ...interface{}) {
		message := fmt.Sprintf(format, v...)
		log.Printf("["+logPrefix+"]: %s\n", message)
	}
}
