package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/yumikokawaii/sherry-archive/serve"
)

func main() {
	rootCmd := &cobra.Command{Use: "sherry-archive"}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		Run:   serve.Server,
	})

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
