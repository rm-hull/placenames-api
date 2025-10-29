package main

import (
	"github.com/rm-hull/placenames-api/cmd"
	"github.com/spf13/cobra"
)

func main() {
	var filePath string
	var port int
	var debug bool
	var topK int
	var numWorkers int

	rootCmd := &cobra.Command{
		Use:  "placenames",
		Long: `Place names auto-suggest API`,
	}
	rootCmd.PersistentFlags().StringVar(&filePath, "file", "./data/placenames_with_relevancy.csv.gz", "Path to place names data file")

	apiServerCmd := &cobra.Command{
		Use:   "api-server [--file <path>] [--port <port>] [--debug] [--top-k <k>]",
		Short: "Start HTTP API server",
		RunE: func(_ *cobra.Command, _ []string) error {
			return cmd.ApiServer(filePath, port, debug, topK)
		},
	}
	apiServerCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")
	apiServerCmd.Flags().IntVar(&topK, "top-k", 100, "Number of top results to store per prefix node")
	apiServerCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging (pprof) - WARING: do not enable in production")

	regenCsvCmd := &cobra.Command{
		Use: "regen-csv [--file <path>] [--num-workers <n>]",
		Short: "Regenerate CSV file scores",
		RunE: func(_ *cobra.Command, _ []string) error {
			return cmd.RegenCSV(filePath, numWorkers)
		},
	}
	regenCsvCmd.Flags().IntVar(&numWorkers, "num-workers", 6, "Number of workers to run concurrently")
	rootCmd.AddCommand(apiServerCmd)
	rootCmd.AddCommand(regenCsvCmd)

	_ = rootCmd.Execute()
}
