package main

import (
	"github.com/spf13/cobra"
	"github.com/yumikokawaii/sherry-archive/jobs"
	"github.com/yumikokawaii/sherry-archive/migrate"
	"github.com/yumikokawaii/sherry-archive/pkg/logger"
	"github.com/yumikokawaii/sherry-archive/serve"
	"go.uber.org/zap"
)

func main() {
	flush := logger.Init()
	defer flush()

	cmd := &cobra.Command{Use: "sherry-archive"}

	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		Run:   serve.Server,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "migrate up",
		Short: "Apply pending database migrations",
		Run:   migrate.Up,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "aggregate-user-interests",
		Short: "Aggregate user interest profiles from tracking events",
		Run:   jobs.Run,
	})

	if err := cmd.Execute(); err != nil {
		zap.L().Fatal("command failed", zap.Error(err))
	}
}
