package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/0xjuanma/golazo/internal/app"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags
var Version = "dev"

var mockFlag bool
var updateFlag bool
var versionFlag bool

var rootCmd = &cobra.Command{
	Use:   "golazo",
	Short: "Football match stats and updates in your terminal",
	Long:  `A modern terminal user interface for real-time football stats and scores, covering multiple leagues and competitions.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			printVersion()
			return
		}

		if updateFlag {
			runUpdate()
			return
		}

		p := tea.NewProgram(app.New(mockFlag), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
			os.Exit(1)
		}
	},
}

// printVersion displays the ASCII logo with gradient and version.
func printVersion() {
	// Render ASCII title with gradient (same as main view)
	title := ui.RenderGradientText(constants.ASCIITitle)

	// Render version with gradient color (use the end color - red)
	endColor, _ := colorful.Hex(constants.GradientEndColor)
	versionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(endColor.Hex()))
	versionText := versionStyle.Render(Version)

	// Concatenate version after the last line of ASCII art
	fmt.Println(title + "" + versionText)
}

// runUpdate executes the install script to update golazo to the latest version.
func runUpdate() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", "irm https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.ps1 | iex")
	} else {
		cmd = exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.sh | bash")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}
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
	rootCmd.Flags().BoolVarP(&updateFlag, "update", "u", false, "Update golazo to the latest version")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
}
