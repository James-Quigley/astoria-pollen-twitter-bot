package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/James-Quigley/astoria-pollen-twitter-bot/internal"
)

func HandleRequest() error {
	return internal.Handle()
}

func main() {
	lambda.Start(HandleRequest)
}
