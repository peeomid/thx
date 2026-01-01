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
		Use:   "areas",
		Short: "List areas",
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
		Use:   "tags",
		Short: "List tags",
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
		Args:  cobra.ExactArgs(1),
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
		Args:  cobra.MaximumNArgs(1),
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
