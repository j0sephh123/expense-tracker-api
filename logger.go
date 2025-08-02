package main

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	warningLogger *log.Logger
	debugLogger   *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		infoLogger:    log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		errorLogger:   log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime),
		warningLogger: log.New(os.Stdout, "[WARNING] ", log.Ldate|log.Ltime),
		debugLogger:   log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime),
	}
}


func (l *Logger) Info(message string) {
	l.infoLogger.Println(message)
	fmt.Printf("[INFO] %s\n", message)
}

func (l *Logger) Error(message string) {
	l.errorLogger.Println(message)
	fmt.Printf("[ERROR] %s\n", message)
}

func (l *Logger) Warning(message string) {
	l.warningLogger.Println(message)
	fmt.Printf("[WARNING] %s\n", message)
}

func (l *Logger) Debug(message string) {
	l.debugLogger.Println(message)
	fmt.Printf("[DEBUG] %s\n", message)
}
