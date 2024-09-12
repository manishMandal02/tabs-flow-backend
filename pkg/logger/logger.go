package logger

import "fmt"

func Error(msg string, err error) {
	fmt.Println("❌ Logger: %v\n [Error]: %v ", msg, err)
}

func Dev(msg interface{}) {
	fmt.Println("🛜 Logger:Dev:", msg)
}
