package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:   "media",
	Short: "media",
	Long:  "media",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCommandImplementation()
	},
}

func rootCommandImplementation() error {
	var logo []string = []string{
		"Stormic media worker",
	}
	for _, l := range logo {
		fmt.Println(l)
	}

	return nil
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
