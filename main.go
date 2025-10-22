package main

import (
	"github.com/rm-hull/placenames-api/cmd"
	"github.com/spf13/cobra"
)

func main() {
	var filePath string
	var port int
	var debug bool

	rootCmd := &cobra.Command{
		Use:  "placenames",
		Long: `Place names auto-suggest API`,
	}

	apiServerCmd := &cobra.Command{
		Use:   "api-server [--file <path>] [--port <port>] [--debug]",
		Short: "Start HTTP API server",
		RunE: func(_ *cobra.Command, _ []string) error {
			return cmd.ApiServer(filePath, port, debug)
		},
	}
	apiServerCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")
	apiServerCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging (pprof) - WARING: do not enable in production")
	apiServerCmd.PersistentFlags().StringVar(&filePath, "file", "./data/placenames_with_relevancy.csv.gz", "Path to place names data file")

	rootCmd.AddCommand(apiServerCmd)

	rootCmd.Execute()
}
