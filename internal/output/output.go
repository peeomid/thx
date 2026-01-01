package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"thx/internal/things"
)

// Mode is output mode.
type Mode string

const (
	ModeDefault Mode = "default"
	ModeJSON    Mode = "json"
	ModeQuiet   Mode = "quiet"
)

// ResolveMode picks output mode from flags/config.
func ResolveMode(jsonFlag, quietFlag bool, cfgFormat string) Mode {
	if jsonFlag {
		return ModeJSON
	}
	if quietFlag {
		return ModeQuiet
	}
	switch strings.ToLower(cfgFormat) {
	case string(ModeJSON):
		return ModeJSON
	case string(ModeQuiet):
		return ModeQuiet
	default:
		return ModeDefault
	}
}

// WriteTodos prints a list of todos.
func WriteTodos(w io.Writer, mode Mode, todos []things.Todo) error {
	switch mode {
	case ModeJSON:
		return writeJSON(w, todos)
	case ModeQuiet:
		for _, todo := range todos {
			if _, err := fmt.Fprintln(w, todo.ID); err != nil {
				return err
			}
		}
		return nil
	default:
		for _, todo := range todos {
			if _, err := fmt.Fprintln(w, formatTodoLine(todo)); err != nil {
				return err
			}
		}
		return nil
	}
}

// WriteProjects prints a list of projects.
func WriteProjects(w io.Writer, mode Mode, projects []things.Project) error {
	switch mode {
	case ModeJSON:
		return writeJSON(w, projects)
	case ModeQuiet:
		for _, project := range projects {
			if _, err := fmt.Fprintln(w, project.ID); err != nil {
				return err
			}
		}
		return nil
	default:
		for _, project := range projects {
			line := formatProjectLine(project)
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
		return nil
	}
}

// WriteAreas prints a list of areas.
func WriteAreas(w io.Writer, mode Mode, areas []things.Area) error {
	switch mode {
	case ModeJSON:
		return writeJSON(w, areas)
	case ModeQuiet:
		for _, area := range areas {
			if _, err := fmt.Fprintln(w, area.ID); err != nil {
				return err
			}
		}
		return nil
	default:
		for _, area := range areas {
			if _, err := fmt.Fprintf(w, "%s\n", area.Title); err != nil {
				return err
			}
		}
		return nil
	}
}

// WriteTags prints a list of tags.
func WriteTags(w io.Writer, mode Mode, tags []things.Tag) error {
	switch mode {
	case ModeJSON:
		return writeJSON(w, tags)
	case ModeQuiet:
		for _, tag := range tags {
			if _, err := fmt.Fprintln(w, tag.ID); err != nil {
				return err
			}
		}
		return nil
	default:
		for _, tag := range tags {
			if _, err := fmt.Fprintf(w, "%s\n", tag.Title); err != nil {
				return err
			}
		}
		return nil
	}
}

// WriteTodoDetail prints a single todo in detail.
func WriteTodoDetail(w io.Writer, mode Mode, todo things.Todo) error {
	switch mode {
	case ModeJSON:
		return writeJSON(w, todo)
	case ModeQuiet:
		_, err := fmt.Fprintln(w, todo.ID)
		return err
	default:
		lines := []string{
			"ID: " + todo.ID,
			"Title: " + todo.Title,
			"Status: " + todo.Status,
			"When: " + todo.When,
		}
		if todo.ProjectTitle != "" {
			lines = append(lines, "Project: "+todo.ProjectTitle)
		}
		if todo.AreaTitle != "" {
			lines = append(lines, "Area: "+todo.AreaTitle)
		}
		if todo.Deadline != "" {
			lines = append(lines, "Deadline: "+todo.Deadline)
		}
		if len(todo.Tags) > 0 {
			sorted := append([]string(nil), todo.Tags...)
			sort.Strings(sorted)
			lines = append(lines, "Tags: "+strings.Join(sorted, ", "))
		}
		if todo.Notes != "" {
			lines = append(lines, "Notes: "+todo.Notes)
		}
		if len(todo.Checklist) > 0 {
			lines = append(lines, "Checklist:")
			for _, item := range todo.Checklist {
				mark := "[ ]"
				if item.Completed {
					mark = "[x]"
				}
				lines = append(lines, fmt.Sprintf("  %s %s", mark, item.Title))
			}
		}
		for _, line := range lines {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
		return nil
	}
}

// WriteProjectDetail prints a project with tasks.
func WriteProjectDetail(w io.Writer, mode Mode, project things.Project) error {
	switch mode {
	case ModeJSON:
		return writeJSON(w, project)
	case ModeQuiet:
		_, err := fmt.Fprintln(w, project.ID)
		return err
	default:
		lines := []string{
			"ID: " + project.ID,
			"Title: " + project.Title,
			"Status: " + project.Status,
			"When: " + project.When,
		}
		if project.AreaTitle != "" {
			lines = append(lines, "Area: "+project.AreaTitle)
		}
		if project.Deadline != "" {
			lines = append(lines, "Deadline: "+project.Deadline)
		}
		if len(project.Tags) > 0 {
			sorted := append([]string(nil), project.Tags...)
			sort.Strings(sorted)
			lines = append(lines, "Tags: "+strings.Join(sorted, ", "))
		}
		if project.Notes != "" {
			lines = append(lines, "Notes: "+project.Notes)
		}
		for _, line := range lines {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
		if len(project.Headings) > 0 {
			if _, err := fmt.Fprintln(w, "Headings:"); err != nil {
				return err
			}
			for _, heading := range project.Headings {
				if _, err := fmt.Fprintln(w, "  - "+heading.Title); err != nil {
					return err
				}
			}
		}
		if len(project.Todos) > 0 {
			if _, err := fmt.Fprintln(w, "Todos:"); err != nil {
				return err
			}
			for _, todo := range project.Todos {
				if _, err := fmt.Fprintln(w, "  - "+todo.Title); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func formatTodoLine(todo things.Todo) string {
	status := "[ ]"
	switch todo.Status {
	case "completed":
		status = "[x]"
	case "canceled":
		status = "[-]"
	}
	line := status + " " + todo.Title
	if len(todo.Tags) > 0 {
		sorted := append([]string(nil), todo.Tags...)
		sort.Strings(sorted)
		line += " [" + strings.Join(sorted, ",") + "]"
	}
	if todo.Deadline != "" {
		line += " due: " + todo.Deadline
	}
	if todo.ProjectTitle != "" {
		line += " (" + todo.ProjectTitle + ")"
	}
	return line
}

func formatProjectLine(project things.Project) string {
	line := project.Title
	if project.AreaTitle != "" {
		line += " (" + project.AreaTitle + ")"
	}
	return line
}

func writeJSON(w io.Writer, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
