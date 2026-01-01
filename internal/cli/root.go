package cli

import (
	"errors"
	"fmt"
	"os"

	"thx/internal/config"
	"thx/internal/db"
	"thx/internal/output"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "thx",
		Short: "thx - Things CLI",
		Long: `thx is a fast, scriptable CLI for Things 3 on macOS.

Read operations query the Things SQLite database directly (read-only).
Write operations use the Things URL scheme.

Prerequisites:
- macOS with Things 3 installed and opened at least once
- Things → Settings → General → Enable Things URLs
- Terminal app granted Full Disk Access (System Settings → Privacy & Security)

Database path resolution order:
1) --database flag
2) config file "database"
3) THINGSDB env var
4) ~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite
5) ~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite

Config file (YAML):
~/.config/thx/config.yaml
Environment overrides use the THX_ prefix, for example:
  THX_FORMAT=json
  THX_DATABASE=/path/to/main.sqlite
  THX_DEFAULTS_WHEN=today
  THX_DEFAULTS_TAGS=work,home`,
		Example: `  thx today
  thx add "Buy milk" --when today --tags errands
  thx search "project alpha" --tag work
  thx show <id> --include-checklist
  thx update <id> --title "New title" --when tomorrow
  thx done <id>`,
	}
	cfgPath   string
	dbPath    string
	format    string
	jsonFlag  bool
	quietFlag bool

	appConfig config.Config
	outMode   output.Mode

	version = "dev"
)

// Execute runs the CLI.
func Execute() error {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "path to config file")
	rootCmd.PersistentFlags().StringVar(&dbPath, "database", "", "path to Things database")
	rootCmd.PersistentFlags().StringVar(&format, "format", "", "output format: default|json|quiet")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output JSON")
	rootCmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "output IDs only")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}
		appConfig = cfg
		if format != "" {
			appConfig.Format = format
		}
		outMode = output.ResolveMode(jsonFlag, quietFlag, appConfig.Format)
		return nil
	}

	rootCmd.SetHelpTemplate(helpTemplate())

	rootCmd.AddCommand(
		newInboxCmd(),
		newTodayCmd(),
		newUpcomingCmd(),
		newAnytimeCmd(),
		newSomedayCmd(),
		newLogbookCmd(),
		newProjectsCmd(),
		newAreasCmd(),
		newTagsCmd(),
		newShowCmd(),
		newSearchCmd(),
		newAddCmd(),
		newAddProjectCmd(),
		newDoneCmd(),
		newCancelCmd(),
		newUpdateCmd(),
		newMoveCmd(),
		newOpenCmd(),
		newQuickCmd(),
		newDoctorCmd(),
		newVersionCmd(),
	)

	return rootCmd.Execute()
}

func helpTemplate() string {
	return `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

Usage:
  {{.UseLine}}

Examples:
{{with .Example}}{{. | trimTrailingWhitespaces}}{{end}}

Commands:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}

Flags:
{{.Flags.FlagUsages | trimTrailingWhitespaces}}

Global Flags:
{{.PersistentFlags.FlagUsages | trimTrailingWhitespaces}}

Output Formats:
  --format default|json|quiet
  --json / --quiet override --format

Config:
  File: ~/.config/thx/config.yaml
  Env:  THX_FORMAT, THX_DATABASE, THX_DEFAULTS_WHEN, THX_DEFAULTS_TAGS

Use "thx help <command>" for more details about a command.
`
}

func openStore() (*db.Store, error) {
	path, err := db.ResolveDatabasePath(dbPath, appConfig.Database)
	if err != nil {
		if errors.Is(err, db.ErrDatabaseNotFound) {
			return nil, fmt.Errorf("database not found; set --database or grant Full Disk Access")
		}
		return nil, err
	}
	store, err := db.OpenStore(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return store, nil
}

func outputError(err error) error {
	if err == nil {
		return nil
	}
	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	return err
}
