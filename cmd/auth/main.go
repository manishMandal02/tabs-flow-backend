package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/manishMandal02/tabsflow-backend/config"
)

func main() {

	// load config
	config.Init()

	fmt.Println("Auth Lambda called!, evn:", config.ENVIRONMENT)

	mux := http.NewServeMux()

	mux.HandlerFunc("/", )

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Auth Lambda called!")
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]string{"message": "Hello World!"})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})

	if config.ENVIRONMENT != "lambda" {
		fmt.Println("Running in dev mode")
		http.ListenAndServe(":8080", nil)
	} else {
		lambda.Start(httpadapter.New(mux).ProxyWithContext)
	}

}
