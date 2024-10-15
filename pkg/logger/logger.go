package logger

import "fmt"

func Error(msg string, err error) {
	fmt.Printf("❌ Logger: %v\n [Error]: %v\n", msg, err)
}

func Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	fmt.Println("❌ Logger:Error: ", msg)
}

func Dev(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	fmt.Println("🛜  Logger:Dev:", msg)
}

func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	fmt.Println("ℹ️  Logger:Info:", msg)
}
