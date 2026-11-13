package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flanksource/gomplate/v4/playground"
)

func evaluateJSON(input string) string {
	var request playground.Request
	var response *playground.Response
	err := json.Unmarshal([]byte(input), &request)
	if err == nil {
		response, err = (&playground.Handler{}).Evaluate(context.Background(), request)
	}
	if err != nil {
		response = &playground.Response{Error: &playground.EvalError{Message: err.Error()}}
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		encoded, err = json.Marshal(playground.Response{Error: &playground.EvalError{
			Message: fmt.Sprintf("encode evaluation result: %s", err),
		}})
		if err != nil {
			panic(err)
		}
	}
	return string(encoded)
}
