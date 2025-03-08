package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "qa-cli",
		Usage: "A CLI for managing QA test runs",
		Action: func(c *cli.Context) error {
			fmt.Println("Hello from the CLI!")
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
