package logger

import "fmt"

func Error(msg string, err error) {
	fmt.Printf("❌ Log: %v\n [Error]: %v\n", msg, err)
}

func Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	fmt.Println("❌ Log:Error: ", msg)
}

func Dev(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	fmt.Println("🛜  Log:Dev:", msg)
}

func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	fmt.Println("ℹ️  Log:Info:", msg)
}
