package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nanolambda",
	Short: "NanoLambda - Intelligent Serverless Platform",
	Long:  `a research-grade serverless platform with predictive ai scaling.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("nanolambda v0.1.0 (research preview)")
		},
	}
	rootCmd.AddCommand(versionCmd)
}