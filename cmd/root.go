package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/0xjuanma/golazo/internal/app"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags
var Version = "dev"

var mockFlag bool
var updateFlag bool
var versionFlag bool
var debugFlag bool
var wcYearFlag string

var rootCmd = &cobra.Command{
	Use:   "golazo",
	Short: "The beautiful game in your terminal",
	Long:  `A minimal TUI for following football matches in real-time. Get live match updates, finished match statistics, and minute-by-minute events directly in your terminal.`,
	// Suppress cobra's default plain-text error printing so we can wrap
	// flag/subcommand errors in the agent-facing JSON envelope (see Execute).
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			version.Print(Version)
			return
		}

		if updateFlag {
			runUpdate()
			return
		}

		// Determine banner conditions
		isDevBuild := Version == "dev"
		newVersionAvailable := false
		storedLatestVersion := ""

		if !isDevBuild {
			if storedLatestVersion, err := data.LoadLatestVersion(); err == nil && storedLatestVersion != "" {
				// Check if new version is available (current app < stored latest)
				newVersionAvailable = version.IsOlder(Version, storedLatestVersion)
			}
		}

		// Check for updates in background (non-blocking)
		go func() {
			// Check immediately if current version is older than stored, OR do daily check
			shouldCheck := data.ShouldCheckVersion()
			if !shouldCheck && storedLatestVersion != "" && !isDevBuild {
				shouldCheck = version.IsOlder(Version, storedLatestVersion)
			}

			if shouldCheck {
				if fetchedVersion, err := data.CheckLatestVersion(); err == nil {
					_ = data.SaveLatestVersion(fetchedVersion)
				}
			}
		}()

		p := tea.NewProgram(app.New(mockFlag, debugFlag, isDevBuild, newVersionAvailable, Version, wcYearFlag), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
			os.Exit(1)
		}
	},
}

// runUpdate executes the appropriate update method based on installation detection.
// It first checks whether the running binary is already on the latest release (or
// is a dev build) and short-circuits with an informative message in that case.
func runUpdate() {
	latest, fetchErr := data.CheckLatestVersion()
	proceed, message := decideUpdate(Version, latest, fetchErr)
	if message != "" {
		fmt.Println(message)
	}
	if !proceed {
		return
	}

	installMethod := detectInstallationMethod()

	switch installMethod {
	case "homebrew":
		fmt.Println("Updating via Homebrew...")
		if err := runBrewUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "Homebrew update failed: %v\n", err)
			fmt.Println("Falling back to install script...")
			if err := runScriptUpdate(); err != nil {
				fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
				os.Exit(1)
			}
		}
	default: // "script"
		fmt.Println("Updating via install script...")
		if err := runScriptUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}
	}
}

// decideUpdate returns whether runUpdate should shell out to the installer, plus
// a user-facing message to print first. It is pure (no I/O) so the policy is unit
// testable without invoking brew or the install script.
//
// Policy:
//   - Dev builds short-circuit (cannot meaningfully self-update).
//   - On fetch failure we proceed: the user explicitly asked to update, so a
//     network flake shouldn't block them — fall back to today's behavior.
//   - If the current version is not older than latest (equal, or local ahead of
//     the published release), short-circuit with an informative message.
//   - Otherwise proceed.
func decideUpdate(current, latest string, fetchErr error) (proceed bool, message string) {
	if current == "dev" {
		return false, "Running a dev build; skipping update."
	}
	if fetchErr != nil || latest == "" {
		return true, ""
	}
	if version.IsOlder(current, latest) {
		return true, ""
	}
	return false, fmt.Sprintf("Already on the latest version (%s).", current)
}

// runBrewUpdate attempts to update golazo via Homebrew.
func runBrewUpdate() error {
	cmd := exec.Command("brew", "upgrade", "0xjuanma/tap/golazo")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		// brew upgrade can exit non-zero for two recoverable reasons:
		//   1. The brew link step failed because a direct binary (e.g. from a
		//      prior script install) already exists at /usr/local/bin/golazo,
		//      preventing Homebrew from creating its symlink.
		//   2. An unrelated brew cleanup error (e.g. Docker CLI plugins
		//      permissions) fires after a successful upgrade+link.
		// In both cases the formula was built successfully; attempt a forced
		// re-link before giving up and falling back to the script.
		fmt.Println("Attempting brew link recovery...")
		linkCmd := exec.Command("brew", "link", "--overwrite", "0xjuanma/tap/golazo")
		linkCmd.Stdout = os.Stdout
		linkCmd.Stderr = os.Stderr
		if linkErr := linkCmd.Run(); linkErr == nil {
			return nil
		}
		return err
	}
	return nil
}

// runScriptUpdate updates golazo via the install script.
func runScriptUpdate() error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", "irm https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.ps1 | iex")
	} else {
		cmd = exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.sh | bash")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// detectInstallationMethod returns "homebrew" or "script" based on how golazo was installed.
func detectInstallationMethod() string {
	// 1. Fast path: check if binary is in Homebrew Cellar
	if isBinaryInCellar() {
		return "homebrew"
	}

	// 2. Fallback: ask brew directly if package is installed
	if isListedInBrew() {
		return "homebrew"
	}

	// 3. Default to script installation
	return "script"
}

// isBinaryInCellar checks if the golazo binary is located in Homebrew's Cellar directory.
func isBinaryInCellar() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}

	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return false
	}

	return strings.Contains(realPath, "/Cellar/golazo/")
}

// isListedInBrew checks if golazo appears in brew's installed package list.
func isListedInBrew() bool {
	if _, err := exec.LookPath("brew"); err != nil {
		return false
	}

	cmd := exec.Command("brew", "list", "golazo")
	return cmd.Run() == nil
}

// Execute runs the root command.
//
// All subcommands call os.Exit() directly with the documented exit codes,
// so any error reaching here is a cobra-level parse failure (unknown command,
// unknown flag, invalid value). We surface those through the same JSON
// envelope used by every other CLI failure so agents can rely on a single
// error contract.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(WriteError(os.Stderr, ErrCodeInvalidArgs, err))
	}
}

func init() {
	rootCmd.Flags().BoolVar(&mockFlag, "mock", false, "Use mock data for all views instead of real API data")
	rootCmd.Flags().BoolVar(&debugFlag, "debug", false, "Enable debug logging to "+data.DebugLogPath())
	rootCmd.Flags().BoolVarP(&updateFlag, "update", "u", false, "Update golazo to the latest version")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	rootCmd.Flags().StringVar(&wcYearFlag, "wc-year", "", "World Cup year to display (e.g. 2026). With --mock uses bundled preview data; without --mock fetches from API.")
}
