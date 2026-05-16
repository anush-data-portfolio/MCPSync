package main

import (
	"mcpsync/cmd"
	_ "mcpsync/internal/agent"
)

func main() {
	cmd.Execute()
}
