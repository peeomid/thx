package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"thx/internal/things"

	_ "modernc.org/sqlite"
)

const (
	statusOpen     = 0
	statusCanceled = 2
	statusDone     = 3
	startInbox     = 0
	startScheduled = 1
	startSomeday   = 2
	typeTodo       = 0
	typeProject    = 1
	typeHeading    = 2
)

// Store provides read access to the Things database.
type Store struct {
	db  *sql.DB
	loc *time.Location
}

func baseTaskFilters(status, taskType int) (string, []any) {
	where := "t.rt1_recurrenceRule IS NULL AND t.status = ? AND t.trashed = 0 AND t.type = ? AND NOT IFNULL(p.trashed, 0) AND NOT IFNULL(hp.trashed, 0)"
	args := []any{status, taskType}
	return where, args
}

// OpenStore opens the Things database in read-only mode.
func OpenStore(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &Store{db: db, loc: time.Local}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// URLSchemeToken returns the Things URL scheme authentication token.
func (s *Store) URLSchemeToken() (string, error) {
	if s == nil || s.db == nil {
		return "", errors.New("database not initialized")
	}
	var token sql.NullString
	err := s.db.QueryRow("SELECT uriSchemeAuthenticationToken FROM TMSettings WHERE uuid = 'RhAzEf6qDxCD5PmnZVtBZR'").Scan(&token)
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", nil
	}
	return token.String, nil
}

// ListInbox returns inbox todos.
func (s *Store) ListInbox(limit, offset int) ([]things.Todo, error) {
	where, args := baseTaskFilters(statusOpen, typeTodo)
	where += " AND t.start = ?"
	args = append(args, startInbox)
	return s.listTodos(where, args, "t.\"index\" ASC", limit, offset)
}

// ListToday returns today todos.
func (s *Store) ListToday(limit, offset int) ([]things.Todo, error) {
	today := int64(things.TimeToThingsDate(time.Now()))
	whereBase, argsBase := baseTaskFilters(statusOpen, typeTodo)

	regularWhere := whereBase + " AND t.start = ? AND t.startDate IS NOT NULL"
	regularArgs := append(append([]any{}, argsBase...), startScheduled)
	regular, err := s.queryTasks(regularWhere, regularArgs, "t.todayIndex ASC", 0, 0)
	if err != nil {
		return nil, err
	}

	unconfirmedScheduledWhere := whereBase + " AND t.start = ? AND t.startDate <= ?"
	unconfirmedScheduledArgs := append(append([]any{}, argsBase...), startSomeday, today)
	unconfirmedScheduled, err := s.queryTasks(unconfirmedScheduledWhere, unconfirmedScheduledArgs, "t.todayIndex ASC", 0, 0)
	if err != nil {
		return nil, err
	}

	unconfirmedOverdueWhere := whereBase + " AND t.startDate IS NULL AND t.deadline <= ? AND t.deadlineSuppressionDate IS NULL"
	unconfirmedOverdueArgs := append(append([]any{}, argsBase...), today)
	unconfirmedOverdue, err := s.queryTasks(unconfirmedOverdueWhere, unconfirmedOverdueArgs, "t.todayIndex ASC", 0, 0)
	if err != nil {
		return nil, err
	}

	rows := append(regular, unconfirmedScheduled...)
	rows = append(rows, unconfirmedOverdue...)
	sort.Slice(rows, func(i, j int) bool {
		leftIndex := int64(1<<62)
		rightIndex := int64(1<<62)
		if rows[i].TodayIndex.Valid {
			leftIndex = rows[i].TodayIndex.Int64
		}
		if rows[j].TodayIndex.Valid {
			rightIndex = rows[j].TodayIndex.Int64
		}
		if leftIndex != rightIndex {
			return leftIndex < rightIndex
		}

		leftStart := int64(1<<62)
		rightStart := int64(1<<62)
		if rows[i].StartDate.Valid {
			leftStart = rows[i].StartDate.Int64
		}
		if rows[j].StartDate.Valid {
			rightStart = rows[j].StartDate.Int64
		}
		return leftStart < rightStart
	})

	if offset > 0 {
		if offset >= len(rows) {
			return []things.Todo{}, nil
		}
		rows = rows[offset:]
	}
	if limit > 0 && limit < len(rows) {
		rows = rows[:limit]
	}

	todos := make([]things.Todo, 0, len(rows))
	for _, row := range rows {
		todos = append(todos, row.toTodo())
	}
	return todos, nil
}

// ListAnytime returns anytime todos.
func (s *Store) ListAnytime(limit, offset int) ([]things.Todo, error) {
	where, args := baseTaskFilters(statusOpen, typeTodo)
	where += " AND t.start = ?"
	args = append(args, startScheduled)
	return s.listTodos(where, args, "t.\"index\" ASC", limit, offset)
}

// ListSomeday returns someday todos.
func (s *Store) ListSomeday(limit, offset int) ([]things.Todo, error) {
	where, args := baseTaskFilters(statusOpen, typeTodo)
	where += " AND t.start = ? AND t.startDate IS NULL"
	args = append(args, startSomeday)
	return s.listTodos(where, args, "t.\"index\" ASC", limit, offset)
}

// ListUpcoming returns upcoming todos within the next rangeDays.
func (s *Store) ListUpcoming(rangeDays, limit, offset int) ([]things.Todo, error) {
	where, args := baseTaskFilters(statusOpen, typeTodo)
	today := int64(things.TimeToThingsDate(time.Now()))
	where += " AND t.start = ? AND t.startDate > ?"
	args = append(args, startSomeday, today)
	if rangeDays > 0 {
		end := int64(things.AddDays(things.TimeToThingsDate(time.Now()), rangeDays, s.loc))
		where += " AND t.startDate <= ?"
		args = append(args, end)
	}
	return s.listTodos(where, args, "t.\"index\" ASC", limit, offset)
}

// ListLogbook returns completed and canceled todos.
func (s *Store) ListLogbook(limit, offset int) ([]things.Todo, error) {
	where, args := baseTaskFilters(statusDone, typeTodo)
	where = strings.Replace(where, "t.status = ?", "t.status IN (?, ?)", 1)
	args = []any{statusDone, statusCanceled, typeTodo}
	return s.listTodos(where, args, "t.stopDate DESC", limit, offset)
}

// ListProjects returns open projects.
func (s *Store) ListProjects(limit, offset int) ([]things.Project, error) {
	where, args := baseTaskFilters(statusOpen, typeProject)
	rows, err := s.queryTasks(where, args, "t.\"index\" ASC", limit, offset)
	if err != nil {
		return nil, err
	}
	projects := make([]things.Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, row.toProject())
	}
	return projects, nil
}

// ListAreas returns areas.
func (s *Store) ListAreas() ([]things.Area, error) {
	rows, err := s.db.Query("SELECT uuid, title FROM TMArea ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	areas := []things.Area{}
	for rows.Next() {
		var id, title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		areas = append(areas, things.Area{ID: id, Title: title})
	}
	return areas, rows.Err()
}

// ListTags returns tags.
func (s *Store) ListTags() ([]things.Tag, error) {
	rows, err := s.db.Query("SELECT uuid, title FROM TMTag ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []things.Tag{}
	for rows.Next() {
		var id, title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		tags = append(tags, things.Tag{ID: id, Title: title})
	}
	return tags, rows.Err()
}

// ShowItem returns a todo or project by id or exact title.
func (s *Store) ShowItem(idOrTitle string, includeChecklist bool) (*things.Todo, *things.Project, error) {
	row, err := s.findTaskByIDOrTitle(idOrTitle)
	if err != nil {
		return nil, nil, err
	}
	if row == nil {
		return nil, nil, sql.ErrNoRows
	}
	if row.Type == typeProject {
		project := row.toProject()
		project.Headings, project.Todos, err = s.projectChildren(row.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, &project, nil
	}
	todo := row.toTodo()
	if includeChecklist {
		checklist, err := s.loadChecklist(row.ID)
		if err != nil {
			return nil, nil, err
		}
		todo.Checklist = checklist
	}
	return &todo, nil, nil
}

// Search finds todos by filters.
func (s *Store) Search(filter SearchFilter) ([]things.Todo, error) {
	where, args := buildSearchWhere(filter, s.loc)
	return s.listTodos(where, args, "t.creationDate DESC", filter.Limit, filter.Offset)
}

// SearchFilter defines search filters.
type SearchFilter struct {
	Query    string
	Tag      string
	Project  string
	Area     string
	Status   string
	Deadline string
	Limit    int
	Offset   int
}

func buildSearchWhere(filter SearchFilter, loc *time.Location) (string, []any) {
	where := []string{"t.rt1_recurrenceRule IS NULL", "t.trashed = 0", "NOT IFNULL(p.trashed, 0)", "NOT IFNULL(hp.trashed, 0)", "t.type = ?"}
	args := []any{typeTodo}

	status := strings.ToLower(strings.TrimSpace(filter.Status))
	switch status {
	case "open", "":
		where = append(where, "t.status = ?")
		args = append(args, statusOpen)
	case "done", "completed":
		where = append(where, "t.status = ?")
		args = append(args, statusDone)
	case "canceled", "cancelled":
		where = append(where, "t.status = ?")
		args = append(args, statusCanceled)
	}

	if filter.Query != "" {
		where = append(where, "(t.title LIKE ? OR t.notes LIKE ?)")
		like := "%" + filter.Query + "%"
		args = append(args, like, like)
	}
	if filter.Tag != "" {
		where = append(where, "tag.title = ?")
		args = append(args, filter.Tag)
	}
	if filter.Project != "" {
		where = append(where, "p.title = ?")
		args = append(args, filter.Project)
	}
	if filter.Area != "" {
		where = append(where, "a.title = ?")
		args = append(args, filter.Area)
	}
	if filter.Deadline != "" {
		encoded := encodeDeadline(filter.Deadline, loc)
		if encoded != 0 {
			where = append(where, "t.deadline = ?")
			args = append(args, int64(encoded))
		}
	}
	return strings.Join(where, " AND "), args
}

func encodeDeadline(value string, loc *time.Location) things.ThingsDate {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "today":
		return things.TimeToThingsDate(time.Now())
	case "tomorrow":
		return things.TimeToThingsDate(time.Now().AddDate(0, 0, 1))
	default:
		parsed, err := time.ParseInLocation("2006-01-02", value, loc)
		if err != nil {
			return 0
		}
		return things.TimeToThingsDate(parsed)
	}
}

func (s *Store) listTodos(where string, args []any, order string, limit, offset int) ([]things.Todo, error) {
	rows, err := s.queryTasks(where, args, order, limit, offset)
	if err != nil {
		return nil, err
	}
	todos := make([]things.Todo, 0, len(rows))
	for _, row := range rows {
		todos = append(todos, row.toTodo())
	}
	return todos, nil
}

func (s *Store) projectChildren(projectID string) ([]things.Todo, []things.Todo, error) {
	headings, err := s.listTodos("t.project = ? AND t.type = ? AND t.trashed = 0", []any{projectID, typeHeading}, "t.\"index\" ASC", 0, 0)
	if err != nil {
		return nil, nil, err
	}
	todos, err := s.listTodos("t.project = ? AND t.type = ? AND t.trashed = 0", []any{projectID, typeTodo}, "t.\"index\" ASC", 0, 0)
	if err != nil {
		return nil, nil, err
	}
	return headings, todos, nil
}

func (s *Store) loadChecklist(taskID string) ([]things.ChecklistItem, error) {
	rows, err := s.db.Query("SELECT title, status FROM TMChecklistItem WHERE task = ? ORDER BY \"index\"", taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []things.ChecklistItem{}
	for rows.Next() {
		var title string
		var status int
		if err := rows.Scan(&title, &status); err != nil {
			return nil, err
		}
		items = append(items, things.ChecklistItem{Title: title, Completed: status != 0})
	}
	return items, rows.Err()
}

func (s *Store) findTaskByIDOrTitle(value string) (*taskRow, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("missing id or title")
	}
	rows, err := s.queryTasks("t.uuid = ?", []any{value}, "", 0, 0)
	if err != nil {
		return nil, err
	}
	if len(rows) == 1 {
		return &rows[0], nil
	}
	if len(rows) > 1 {
		return nil, errors.New("multiple items matched id")
	}

	rows, err = s.queryTasks("t.title = ?", []any{value}, "", 0, 0)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	if len(rows) > 1 {
		return nil, errors.New("multiple items matched title")
	}
	return &rows[0], nil
}

func (s *Store) queryTasks(where string, args []any, order string, limit, offset int) ([]taskRow, error) {
	base := taskSelectBase
	if where != "" {
		base += " WHERE " + where
	}
	base += " GROUP BY t.uuid"
	if order != "" {
		base += " ORDER BY " + order
	}
	if limit > 0 {
		base += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		base += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := s.db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []taskRow{}
	for rows.Next() {
		var row taskRow
		if err := row.scan(rows); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

const taskSelectBase = `SELECT
		t.uuid, t.title, t.status, t.type, t.start, t.startDate, t.deadline, t.trashed,
		t.notes, t.project, t.area, t.heading, t.creationDate, t.userModificationDate, t.stopDate, t.todayIndex,
		p.title as project_title, a.title as area_title,
		GROUP_CONCAT(tag.title) as tags
	FROM TMTask t
	LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
	LEFT JOIN TMTag tag ON tt.tags = tag.uuid
	LEFT JOIN TMTask p ON t.project = p.uuid
	LEFT JOIN TMTask h ON t.heading = h.uuid
	LEFT JOIN TMTask hp ON h.project = hp.uuid
	LEFT JOIN TMArea a ON t.area = a.uuid`

type taskRow struct {
	ID           string
	Title        string
	Status       int
	Type         int
	Start        int
	StartDate    sql.NullInt64
	Deadline     sql.NullInt64
	Trashed      int
	Notes        sql.NullString
	ProjectID    sql.NullString
	AreaID       sql.NullString
	HeadingID    sql.NullString
	Created      sql.NullFloat64
	Modified     sql.NullFloat64
	StopDate     sql.NullFloat64
	TodayIndex   sql.NullInt64
	ProjectTitle sql.NullString
	AreaTitle    sql.NullString
	Tags         sql.NullString
}

func (r *taskRow) scan(rows *sql.Rows) error {
	return rows.Scan(
		&r.ID,
		&r.Title,
		&r.Status,
		&r.Type,
		&r.Start,
		&r.StartDate,
		&r.Deadline,
		&r.Trashed,
		&r.Notes,
		&r.ProjectID,
		&r.AreaID,
		&r.HeadingID,
		&r.Created,
		&r.Modified,
		&r.StopDate,
		&r.TodayIndex,
		&r.ProjectTitle,
		&r.AreaTitle,
		&r.Tags,
	)
}

func (r taskRow) toTodo() things.Todo {
	return things.Todo{
		ID:           r.ID,
		Title:        r.Title,
		Status:       statusToString(r.Status),
		Type:         typeToString(r.Type),
		When:         whenFromRow(r.Start, r.StartDate),
		Deadline:     deadlineFromRow(r.Deadline),
		Tags:         splitTags(r.Tags),
		ProjectID:    nullString(r.ProjectID),
		ProjectTitle: nullString(r.ProjectTitle),
		AreaID:       nullString(r.AreaID),
		AreaTitle:    nullString(r.AreaTitle),
		Notes:        nullString(r.Notes),
		Created:      timeFromFloat(r.Created),
		Modified:     timeFromFloat(r.Modified),
	}
}

func (r taskRow) toProject() things.Project {
	return things.Project{
		ID:        r.ID,
		Title:     r.Title,
		Status:    statusToString(r.Status),
		When:      whenFromRow(r.Start, r.StartDate),
		Deadline:  deadlineFromRow(r.Deadline),
		Tags:      splitTags(r.Tags),
		AreaID:    nullString(r.AreaID),
		AreaTitle: nullString(r.AreaTitle),
		Notes:     nullString(r.Notes),
	}
}

func statusToString(status int) string {
	switch status {
	case statusDone:
		return "completed"
	case statusCanceled:
		return "canceled"
	default:
		return "open"
	}
}

func typeToString(typ int) string {
	switch typ {
	case typeProject:
		return "project"
	case typeHeading:
		return "heading"
	default:
		return "todo"
	}
}

func whenFromRow(start int, startDate sql.NullInt64) string {
	switch start {
	case startInbox:
		return "inbox"
	case startSomeday:
		return "someday"
	case startScheduled:
		if !startDate.Valid {
			return "anytime"
		}
		date := things.ThingsDate(startDate.Int64)
		if date == things.TimeToThingsDate(time.Now()) {
			return "today"
		}
		if date == things.TimeToThingsDate(time.Now().AddDate(0, 0, 1)) {
			return "tomorrow"
		}
		return things.FormatThingsDate(date, time.Local)
	default:
		return ""
	}
}

func deadlineFromRow(deadline sql.NullInt64) string {
	if !deadline.Valid {
		return ""
	}
	return things.FormatThingsDate(things.ThingsDate(deadline.Int64), time.Local)
}

func splitTags(tags sql.NullString) []string {
	if !tags.Valid {
		return nil
	}
	parts := strings.Split(tags.String, ",")
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}
	sort.Strings(clean)
	return clean
}

func nullString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func timeFromFloat(value sql.NullFloat64) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	seconds := int64(value.Float64)
	return time.Unix(seconds, 0)
}
