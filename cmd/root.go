package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jenkinsmaster-cli",
	Short: "CLI tool to deploy JenkinsMaster",
	Long:  `An interactive CLI tool to deploy JenkinsMaster on various platforms.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
