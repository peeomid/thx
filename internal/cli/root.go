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
		newVersionCmd(),
	)

	return rootCmd.Execute()
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
