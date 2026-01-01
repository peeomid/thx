package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"thx/internal/things"

	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var (
		when      string
		deadline  string
		tags      string
		project   string
		area      string
		checklist string
		notes     string
	)
	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a todo",
		Long: `Add a new to-do to Things using the URL scheme.

If --when is not set, the default comes from config (defaults.when).
If --tags is not set, defaults.tags are applied.`,
		Example: `  thx add "Buy milk"
  thx add "Plan trip" --when someday --tags travel,personal
  thx add "Pack bags" --project "Vacation" --checklist "Passport,Toothbrush"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if when == "" {
				when = appConfig.Defaults.When
			}
			if tags == "" && len(appConfig.Defaults.Tags) > 0 {
				tags = strings.Join(appConfig.Defaults.Tags, ",")
			}

			url := things.BuildAddURL(things.AddOptions{
				Title:     args[0],
				When:      when,
				Deadline:  deadline,
				Tags:      parseCommaList(tags),
				Project:   project,
				Area:      area,
				Checklist: parseCommaList(checklist),
				Notes:     notes,
			})
			return executeURL(url)
		},
	}
	cmd.Flags().StringVar(&when, "when", "", "schedule: inbox|today|tomorrow|evening|anytime|someday|YYYY-MM-DD")
	cmd.Flags().StringVar(&deadline, "deadline", "", "deadline date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&tags, "tags", "", "comma-separated tags")
	cmd.Flags().StringVar(&project, "project", "", "project name")
	cmd.Flags().StringVar(&area, "area", "", "area name")
	cmd.Flags().StringVar(&checklist, "checklist", "", "comma-separated checklist items")
	cmd.Flags().StringVar(&notes, "notes", "", "notes")
	return cmd
}

func newAddProjectCmd() *cobra.Command {
	var (
		when     string
		deadline string
		tags     string
		area     string
		todos    string
		notes    string
	)
	cmd := &cobra.Command{
		Use:   "add-project <title>",
		Short: "Add a project",
		Long:  `Add a new project to Things.`,
		Example: `  thx add-project "Launch v2"
  thx add-project "Move House" --area "Personal" --todos "Book movers,Pack boxes"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := things.BuildAddProjectURL(things.AddProjectOptions{
				Title:    args[0],
				When:     when,
				Deadline: deadline,
				Tags:     parseCommaList(tags),
				Area:     area,
				Todos:    parseCommaList(todos),
				Notes:    notes,
			})
			return executeURL(url)
		},
	}
	cmd.Flags().StringVar(&when, "when", "", "schedule: inbox|today|tomorrow|evening|anytime|someday|YYYY-MM-DD")
	cmd.Flags().StringVar(&deadline, "deadline", "", "deadline date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&tags, "tags", "", "comma-separated tags")
	cmd.Flags().StringVar(&area, "area", "", "area name")
	cmd.Flags().StringVar(&todos, "todos", "", "comma-separated todo titles")
	cmd.Flags().StringVar(&notes, "notes", "", "notes")
	return cmd
}

func newDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "done <id>",
		Short:   "Mark a task complete",
		Long:    `Mark a task as completed by UUID.`,
		Example: `  thx done <id>`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := fetchAuthToken()
			if err != nil {
				return err
			}
			return executeURL(things.BuildDoneURL(args[0], token))
		},
	}
	return cmd
}

func newCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel <id>",
		Short:   "Cancel a task",
		Long:    `Cancel a task by UUID.`,
		Example: `  thx cancel <id>`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := fetchAuthToken()
			if err != nil {
				return err
			}
			return executeURL(things.BuildCancelURL(args[0], token))
		},
	}
	return cmd
}

func newUpdateCmd() *cobra.Command {
	var (
		title    string
		when     string
		deadline string
		tags     string
		project  string
		area     string
		notes    string
	)
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an item",
		Long:  `Update fields on a to-do or project by UUID.`,
		Example: `  thx update <id> --title "New title"
  thx update <id> --when tomorrow --deadline 2025-01-15
  thx update <id> --project "Project Alpha"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" && when == "" && deadline == "" && tags == "" && project == "" && area == "" && notes == "" {
				return errors.New("no update fields provided")
			}
			token, err := fetchAuthToken()
			if err != nil {
				return err
			}
			url := things.BuildUpdateURL(things.UpdateOptions{
				ID:        args[0],
				Title:     title,
				When:      when,
				Deadline:  deadline,
				Tags:      parseCommaList(tags),
				Project:   project,
				Area:      area,
				Notes:     notes,
				AuthToken: token,
			})
			return executeURL(url)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&when, "when", "", "schedule: inbox|today|tomorrow|evening|anytime|someday|YYYY-MM-DD")
	cmd.Flags().StringVar(&deadline, "deadline", "", "deadline date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&tags, "tags", "", "comma-separated tags")
	cmd.Flags().StringVar(&project, "project", "", "project name")
	cmd.Flags().StringVar(&area, "area", "", "area name")
	cmd.Flags().StringVar(&notes, "notes", "", "notes")
	return cmd
}

func newMoveCmd() *cobra.Command {
	var (
		project string
		area    string
		heading string
	)
	cmd := &cobra.Command{
		Use:   "move <id>",
		Short: "Move an item",
		Long:  `Move a to-do into a project or area (optionally under a heading).`,
		Example: `  thx move <id> --to-project "Project Alpha"
  thx move <id> --to-area "Personal"
  thx move <id> --to-project "Project Alpha" --to-heading <heading-id>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" && area == "" && heading == "" {
				return errors.New("no destination provided")
			}
			token, err := fetchAuthToken()
			if err != nil {
				return err
			}
			url := things.BuildMoveURL(things.MoveOptions{
				ID:        args[0],
				Project:   project,
				Area:      area,
				HeadingID: heading,
				AuthToken: token,
			})
			return executeURL(url)
		},
	}
	cmd.Flags().StringVar(&project, "to-project", "", "destination project")
	cmd.Flags().StringVar(&area, "to-area", "", "destination area")
	cmd.Flags().StringVar(&heading, "to-heading", "", "destination heading id")
	return cmd
}

func newOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open <id-or-list>",
		Short: "Open item or list in Things",
		Long: `Open an item or list in Things.
Valid list ids include: inbox, today, upcoming, anytime, someday, logbook.`,
		Example: `  thx open today
  thx open <id>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := things.BuildOpenURL(args[0])
			return executeURL(url)
		},
	}
	return cmd
}

func newQuickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "quick <text>",
		Short:   "Open quick entry with text",
		Long:    `Open the Things Quick Entry window with prefilled text.`,
		Example: `  thx quick "Call dentist tomorrow"`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeURL(things.BuildQuickURL(args[0]))
		},
	}
	return cmd
}

func executeURL(url string) error {
	escaped := strings.ReplaceAll(url, "\"", "\\\"")
	cmd := exec.Command("osascript", "-e", fmt.Sprintf("tell application \"Things3\" to open location \"%s\"", escaped))
	if err := cmd.Run(); err == nil {
		return nil
	}

	fallback := exec.Command("open", url)
	if err := fallback.Run(); err != nil {
		return fmt.Errorf("failed to open Things URL: %w", err)
	}
	return nil
}

func fetchAuthToken() (string, error) {
	store, err := openStore()
	if err != nil {
		return "", outputError(err)
	}
	defer store.Close()

	token, err := store.URLSchemeToken()
	if err != nil {
		return "", outputError(fmt.Errorf("failed to read Things auth token: %w", err))
	}
	if token == "" {
		return "", outputError(errors.New("missing Things URL scheme auth token"))
	}
	return token, nil
}

func parseCommaList(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		items = append(items, trimmed)
	}
	return items
}
