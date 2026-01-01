package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"thx/internal/db"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose configuration and permissions",
		Long: `Run diagnostics for common setup issues:
- Config discovery
- Database path resolution and readability
- Things URL scheme auth token availability`,
		Example: `  thx doctor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var hasError bool
			out := cmd.OutOrStdout()

			fmt.Fprintln(out, "thx doctor")
			fmt.Fprintln(out, "-----------")

			printConfigStatus(out)

			if err := printDatabaseStatus(out); err != nil {
				hasError = true
			}

			if err := printAuthTokenStatus(out); err != nil {
				hasError = true
			}

			if hasError {
				return errors.New("doctor found issues")
			}
			return nil
		},
	}
	return cmd
}

func printConfigStatus(out io.Writer) {
	path := cfgPath
	if path == "" {
		path = expandDefaultConfigPath()
	}
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(out, "Config file: %s (found)\n", path)
		return
	}
	fmt.Fprintf(out, "Config file: %s (not found)\n", path)
	fmt.Fprintln(out, "  Tip: create ~/.config/thx/config.yaml if you want defaults.")
}

func printDatabaseStatus(out io.Writer) error {
	fmt.Fprintln(out, "Database:")
	path, err := db.ResolveDatabasePath(dbPath, appConfig.Database)
	if err != nil {
		fmt.Fprintf(out, "  Resolve: error (%v)\n", err)
		fmt.Fprintln(out, "  Tip: set --database or THINGSDB, or grant Full Disk Access.")
		return err
	}
	fmt.Fprintf(out, "  Resolve: %s\n", path)
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(out, "  Read: error (%v)\n", err)
		fmt.Fprintln(out, "  Tip: grant Full Disk Access to your terminal app.")
		return err
	}
	_ = file.Close()
	fmt.Fprintln(out, "  Read: ok")
	return nil
}

func printAuthTokenStatus(out io.Writer) error {
	fmt.Fprintln(out, "Things URL scheme:")
	store, err := openStore()
	if err != nil {
		fmt.Fprintf(out, "  Open DB: error (%v)\n", err)
		return err
	}
	defer store.Close()

	token, err := store.URLSchemeToken()
	if err != nil {
		fmt.Fprintf(out, "  Auth token: error (%v)\n", err)
		fmt.Fprintln(out, "  Tip: open Things once and enable Things URLs.")
		return err
	}
	if token == "" {
		fmt.Fprintln(out, "  Auth token: missing")
		fmt.Fprintln(out, "  Tip: Things → Settings → General → Enable Things URLs.")
		return errors.New("missing Things URL scheme auth token")
	}
	fmt.Fprintln(out, "  Auth token: ok")
	return nil
}

func expandDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.config/thx/config.yaml"
	}
	return filepath.Join(home, ".config", "thx", "config.yaml")
}
