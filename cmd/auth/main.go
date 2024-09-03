package main

import (
	"context"
	"fmt"
)

type LambdaEvent struct {
	Message string `json:"message"`
}

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	return fmt.Sprintf("Hi %s!", name.Name), nil
}

func main() {
	// lambda.C(HandleRequest)
	HandleRequest(context.Background(), MyEvent{Name: "John Doe"})
}
