package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kraken-trader",
	Short: "Autonomous crypto trading bot",
	Long: `Kraken Trader is an autonomous crypto trading bot that uses AI to make trading decisions.

It integrates with Kraken CLI, PRISM API, Ollama LLM, and various data stores.`,
}

var configPath string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the trading bot",
	Long:  `Start the autonomous crypto trading bot with all its components.`,
	RunE:  runRun,
}

func init() {
	runCmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")
}

// Execute runs the root command
func Execute() error {
	rootCmd.AddCommand(runCmd)
	return rootCmd.Execute()
}
