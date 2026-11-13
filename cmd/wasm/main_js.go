//go:build js && wasm

package main

import "syscall/js"

func main() {
	evaluate := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return `{"result":"","durationMs":0,"error":{"message":"expected one JSON request string"}}`
		}
		return evaluateJSON(args[0].String())
	})
	defer evaluate.Release()
	js.Global().Set("gomplateEvaluate", evaluate)
	js.Global().Call("gomplateReady")
	select {}
}
