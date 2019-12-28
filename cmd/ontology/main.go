package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ontology",
	Short: "Ontology is a cloud-native asset inventory",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var (
	serverURL string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server-url", "s", "ws://127.0.0.1:8182/gremlin", "websocket url to the gremlin server")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
