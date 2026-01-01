package things

import "time"

// ChecklistItem represents a Things checklist item.
type ChecklistItem struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Todo represents a Things task or heading.
type Todo struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Status       string          `json:"status"`
	Type         string          `json:"type"`
	When         string          `json:"when"`
	Deadline     string          `json:"deadline"`
	Tags         []string        `json:"tags"`
	ProjectID    string          `json:"project"`
	ProjectTitle string          `json:"project_title"`
	AreaID       string          `json:"area"`
	AreaTitle    string          `json:"area_title"`
	Notes        string          `json:"notes"`
	Checklist    []ChecklistItem `json:"checklist"`
	Created      time.Time       `json:"created"`
	Modified     time.Time       `json:"modified"`
}

// Project represents a Things project.
type Project struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	When      string   `json:"when"`
	Deadline  string   `json:"deadline"`
	Tags      []string `json:"tags"`
	AreaID    string   `json:"area"`
	AreaTitle string   `json:"area_title"`
	Notes     string   `json:"notes"`
	Todos     []Todo   `json:"todos"`
	Headings  []Todo   `json:"headings"`
}

// Area represents a Things area.
type Area struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Projects []Project `json:"projects"`
	Todos    []Todo    `json:"todos"`
}

// Tag represents a Things tag.
type Tag struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
