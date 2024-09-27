package logger

import "fmt"

func Error(msg string, err error) {
	fmt.Printf("❌ Logger: %v\n [Error]: %v\n", msg, err)
}

func Dev(msg interface{}) {
	fmt.Printf("🛜 Logger:Dev: %v\n", msg)
}
