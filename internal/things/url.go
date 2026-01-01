package things

import (
	"net/url"
	"sort"
	"strings"
)

// AddOptions represents fields for creating a todo.
type AddOptions struct {
	Title     string
	When      string
	Deadline  string
	Tags      []string
	Project   string
	Area      string
	Checklist []string
	Notes     string
}

// AddProjectOptions represents fields for creating a project.
type AddProjectOptions struct {
	Title    string
	When     string
	Deadline string
	Tags     []string
	Area     string
	Todos    []string
	Notes    string
}

// UpdateOptions represents fields for updating a task or project.
type UpdateOptions struct {
	ID       string
	Title    string
	When     string
	Deadline string
	Tags     []string
	Project  string
	Area     string
	Notes    string
	AuthToken string
}

// MoveOptions represents fields for moving items.
type MoveOptions struct {
	ID        string
	Project   string
	Area      string
	HeadingID string
	AuthToken string
}

// BuildAddURL builds a Things URL to add a todo.
func BuildAddURL(opts AddOptions) string {
	values := url.Values{}
	values.Set("title", opts.Title)
	setOptional(values, "when", opts.When)
	setOptional(values, "deadline", opts.Deadline)
	setOptional(values, "list", firstNonEmpty(opts.Project, opts.Area))
	setOptional(values, "notes", opts.Notes)
	setTags(values, opts.Tags)
	setChecklist(values, opts.Checklist)
	return buildURL("add", values)
}

// BuildAddProjectURL builds a Things URL to add a project.
func BuildAddProjectURL(opts AddProjectOptions) string {
	values := url.Values{}
	values.Set("title", opts.Title)
	setOptional(values, "when", opts.When)
	setOptional(values, "deadline", opts.Deadline)
	setOptional(values, "area", opts.Area)
	setOptional(values, "notes", opts.Notes)
	setTags(values, opts.Tags)
	if len(opts.Todos) > 0 {
		values.Set("to-dos", strings.Join(opts.Todos, "\n"))
	}
	return buildURL("add-project", values)
}

// BuildUpdateURL builds a Things URL to update an item.
func BuildUpdateURL(opts UpdateOptions) string {
	values := url.Values{}
	values.Set("id", opts.ID)
	setOptional(values, "title", opts.Title)
	setOptional(values, "when", opts.When)
	setOptional(values, "deadline", opts.Deadline)
	setOptional(values, "list", firstNonEmpty(opts.Project, opts.Area))
	setOptional(values, "notes", opts.Notes)
	setOptional(values, "auth-token", opts.AuthToken)
	setTags(values, opts.Tags)
	return buildURL("update", values)
}

// BuildMoveURL builds a Things URL to move an item.
func BuildMoveURL(opts MoveOptions) string {
	values := url.Values{}
	values.Set("id", opts.ID)
	setOptional(values, "list", firstNonEmpty(opts.Project, opts.Area))
	setOptional(values, "heading-id", opts.HeadingID)
	setOptional(values, "auth-token", opts.AuthToken)
	return buildURL("update", values)
}

// BuildDoneURL builds a Things URL to mark a task complete.
func BuildDoneURL(id, authToken string) string {
	values := url.Values{}
	values.Set("id", id)
	values.Set("completed", "true")
	setOptional(values, "auth-token", authToken)
	return buildURL("update", values)
}

// BuildCancelURL builds a Things URL to cancel a task.
func BuildCancelURL(id, authToken string) string {
	values := url.Values{}
	values.Set("id", id)
	values.Set("canceled", "true")
	setOptional(values, "auth-token", authToken)
	return buildURL("update", values)
}

// BuildOpenURL builds a Things URL to open an item or list.
func BuildOpenURL(target string) string {
	values := url.Values{}
	values.Set("id", target)
	return buildURL("show", values)
}

// BuildQuickURL opens quick entry with a prefilled title.
func BuildQuickURL(text string) string {
	values := url.Values{}
	values.Set("title", text)
	values.Set("show-quick-entry", "true")
	return buildURL("add", values)
}

func buildURL(action string, values url.Values) string {
	encoded := values.Encode()
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	return "things:///" + action + "?" + encoded
}

func setOptional(values url.Values, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	values.Set(key, value)
}

func setTags(values url.Values, tags []string) {
	if len(tags) == 0 {
		return
	}
	clean := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}
	if len(clean) == 0 {
		return
	}
	sort.Strings(clean)
	values.Set("tags", strings.Join(clean, ","))
}

func setChecklist(values url.Values, items []string) {
	if len(items) == 0 {
		return
	}
	clean := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}
	if len(clean) == 0 {
		return
	}
	values.Set("checklist-items", strings.Join(clean, "\n"))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
