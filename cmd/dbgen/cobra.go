package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	outputDir   string
	packageName string
)

var rootCmd = &cobra.Command{
	Use:   "dbgen [-o <outputDir>] [-p <package name>] <schema...>",
	Short: "Generate database access code",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		for _, schemaPath := range args {
			if err := gen(schemaPath); err != nil {
				fmt.Printf("%v", err)
			}
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "./model", "output directory")
	rootCmd.PersistentFlags().StringVarP(&packageName, "package", "p", "model", "package name")
}
