package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
)

var debugCommand = &cobra.Command{
	Use:   "debug",
	Short: "debug",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return debugCommandImpl()
	},
}

func debugCommandImpl() error {
	return nil
}

func init() {
	rootCommand.AddCommand(debugCommand)
}
