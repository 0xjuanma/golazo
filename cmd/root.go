package cmd

import (
	"fmt"
	"os"

	"github.com/0xjuanma/golazo/internal/app"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var mockFlag bool

var rootCmd = &cobra.Command{
	Use:   "golazo",
	Short: "Football match stats and updates in your terminal",
	Long:  `A modern terminal user interface for real-time football stats and scores, covering multiple leagues and competitions.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(app.NewModel(mockFlag), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute runs the root command.
// Errors are written to stderr and the program exits with code 1 on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&mockFlag, "mock", false, "Use mock data for all views instead of real API data")
}
