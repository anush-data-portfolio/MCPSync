package main

import (
	"github.com/anush-data-portfolio/MCPSync/cmd"
	_ "github.com/anush-data-portfolio/MCPSync/internal/agent"
)

func main() {
	cmd.Execute()
}
