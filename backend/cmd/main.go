package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/yumikokawaii/sherry-archive/migrate"
	"github.com/yumikokawaii/sherry-archive/serve"
)

func main() {
	cmd := &cobra.Command{Use: "sherry-archive"}

	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		Run:   serve.Server,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Apply pending database migrations",
		Run:   migrate.Up,
	})

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
