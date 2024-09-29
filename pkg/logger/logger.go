package logger

import "fmt"

func Error(msg string, err error) {
	fmt.Printf("❌ Logger: %v\n [Error]: %v\n", msg, err)
}

func Dev(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	fmt.Println("🛜  Logger:Dev:", msg)
}
