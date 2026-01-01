package cli

import (
	"errors"
	"fmt"

	"thx/internal/db"
	"thx/internal/output"

	"github.com/spf13/cobra"
)

func newInboxCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "Show inbox items",
		Long:  `List all open to-dos in the Inbox.`,
		Example: `  thx inbox
  thx inbox --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListInbox(limit, offset)
			if err != nil {
				return outputError(err)
			}
			return output.WriteTodos(cmd.OutOrStdout(), outMode, items)
		},
	}
	addPagingFlags(cmd, &limit, &offset)
	return cmd
}

func newTodayCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "today",
		Short: "Show today items",
		Long: `List Today items using the same logic as Things:
- Scheduled items with a start date
- Scheduled Someday items that are due
- Overdue deadlines that are not suppressed`,
		Example: `  thx today
  thx today --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListToday(limit, offset)
			if err != nil {
				return outputError(err)
			}
			return output.WriteTodos(cmd.OutOrStdout(), outMode, items)
		},
	}
	addPagingFlags(cmd, &limit, &offset)
	return cmd
}

func newUpcomingCmd() *cobra.Command {
	var limit, offset, rangeDays int
	cmd := &cobra.Command{
		Use:   "upcoming",
		Short: "Show upcoming items",
		Long:  `List scheduled Someday items with a future start date.`,
		Example: `  thx upcoming
  thx upcoming --range-days 30`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListUpcoming(rangeDays, limit, offset)
			if err != nil {
				return outputError(err)
			}
			return output.WriteTodos(cmd.OutOrStdout(), outMode, items)
		},
	}
	cmd.Flags().IntVar(&rangeDays, "range-days", 14, "lookahead range in days")
	addPagingFlags(cmd, &limit, &offset)
	return cmd
}

func newAnytimeCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "anytime",
		Short: "Show anytime items",
		Long:  `List Anytime items (including those with deadlines).`,
		Example: `  thx anytime
  thx anytime --limit 100`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListAnytime(limit, offset)
			if err != nil {
				return outputError(err)
			}
			return output.WriteTodos(cmd.OutOrStdout(), outMode, items)
		},
	}
	addPagingFlags(cmd, &limit, &offset)
	return cmd
}

func newSomedayCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "someday",
		Short: "Show someday items",
		Long:  `List Someday items without a start date.`,
		Example: `  thx someday
  thx someday --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListSomeday(limit, offset)
			if err != nil {
				return outputError(err)
			}
			return output.WriteTodos(cmd.OutOrStdout(), outMode, items)
		},
	}
	addPagingFlags(cmd, &limit, &offset)
	return cmd
}

func newLogbookCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "logbook",
		Short: "Show completed items",
		Long:  `List completed and canceled items, newest first.`,
		Example: `  thx logbook
  thx logbook --limit 25`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListLogbook(limit, offset)
			if err != nil {
				return outputError(err)
			}
			return output.WriteTodos(cmd.OutOrStdout(), outMode, items)
		},
	}
	addPagingFlags(cmd, &limit, &offset)
	return cmd
}

func newProjectsCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List projects",
		Long:  `List all open projects.`,
		Example: `  thx projects
  thx projects --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListProjects(limit, offset)
			if err != nil {
				return outputError(err)
			}
			return output.WriteProjects(cmd.OutOrStdout(), outMode, items)
		},
	}
	addPagingFlags(cmd, &limit, &offset)
	return cmd
}

func newAreasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "areas",
		Short:   "List areas",
		Long:    `List all areas.`,
		Example: `  thx areas`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListAreas()
			if err != nil {
				return outputError(err)
			}
			return output.WriteAreas(cmd.OutOrStdout(), outMode, items)
		},
	}
	return cmd
}

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tags",
		Short:   "List tags",
		Long:    `List all tags.`,
		Example: `  thx tags`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.ListTags()
			if err != nil {
				return outputError(err)
			}
			return output.WriteTags(cmd.OutOrStdout(), outMode, items)
		},
	}
	return cmd
}

func newShowCmd() *cobra.Command {
	var includeChecklist bool
	cmd := &cobra.Command{
		Use:   "show <id-or-title>",
		Short: "Show item details",
		Long: `Show a single to-do or project by UUID or exact title.
For projects, use --include-checklist to include checklist items in nested to-dos.`,
		Example: `  thx show <id>
  thx show "Project Alpha"
  thx show <id> --include-checklist`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			todo, project, err := store.ShowItem(args[0], includeChecklist)
			if err != nil {
				if errors.Is(err, db.ErrDatabaseNotFound) {
					return outputError(err)
				}
				return outputError(err)
			}
			if project != nil {
				return output.WriteProjectDetail(cmd.OutOrStdout(), outMode, *project)
			}
			if todo == nil {
				return outputError(fmt.Errorf("item not found"))
			}
			return output.WriteTodoDetail(cmd.OutOrStdout(), outMode, *todo)
		},
	}
	cmd.Flags().BoolVar(&includeChecklist, "include-checklist", false, "include checklist items")
	return cmd
}

func newSearchCmd() *cobra.Command {
	var filter db.SearchFilter
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search tasks",
		Long:  `Search tasks by title or notes, optionally filtered by tag, project, area, status, or deadline.`,
		Example: `  thx search "roadmap"
  thx search --tag work
  thx search "alpha" --project "Project Alpha"
  thx search --status done
  thx search --deadline today`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				filter.Query = args[0]
			}
			store, err := openStore()
			if err != nil {
				return outputError(err)
			}
			defer store.Close()

			items, err := store.Search(filter)
			if err != nil {
				return outputError(err)
			}
			return output.WriteTodos(cmd.OutOrStdout(), outMode, items)
		},
	}
	cmd.Flags().StringVar(&filter.Tag, "tag", "", "filter by tag")
	cmd.Flags().StringVar(&filter.Project, "project", "", "filter by project")
	cmd.Flags().StringVar(&filter.Area, "area", "", "filter by area")
	cmd.Flags().StringVar(&filter.Status, "status", "open", "filter by status: open|done|canceled")
	cmd.Flags().StringVar(&filter.Deadline, "deadline", "", "filter by deadline (today|tomorrow|YYYY-MM-DD)")
	addPagingFlags(cmd, &filter.Limit, &filter.Offset)
	return cmd
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version)
			return err
		},
	}
	return cmd
}
