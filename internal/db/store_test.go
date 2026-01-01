package db

import (
	"database/sql"
	"testing"
	"time"

	"thx/internal/things"

	_ "modernc.org/sqlite"
)

func TestListAndSearch(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	inbox, err := store.ListInbox(0, 0)
	if err != nil {
		t.Fatalf("ListInbox: %v", err)
	}
	if len(inbox) != 1 || inbox[0].Title != "Inbox Task" {
		t.Fatalf("unexpected inbox results: %+v", inbox)
	}

	today, err := store.ListToday(0, 0)
	if err != nil {
		t.Fatalf("ListToday: %v", err)
	}
	if len(today) != 4 {
		t.Fatalf("unexpected today count: %d", len(today))
	}
	foundToday := map[string]bool{}
	for _, item := range today {
		foundToday[item.Title] = true
	}
	if !foundToday["Today Task"] || !foundToday["Project Task"] || !foundToday["Checklist Task"] || !foundToday["Deadline Today"] {
		t.Fatalf("missing today items: %+v", foundToday)
	}
	if foundToday["Suppressed Deadline"] || foundToday["Recurring Task"] {
		t.Fatalf("unexpected today items: %+v", foundToday)
	}

	upcoming, err := store.ListUpcoming(7, 0, 0)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if len(upcoming) != 1 || upcoming[0].Title != "Upcoming Task" {
		t.Fatalf("unexpected upcoming results: %+v", upcoming)
	}

	anytime, err := store.ListAnytime(0, 0)
	if err != nil {
		t.Fatalf("ListAnytime: %v", err)
	}
	if len(anytime) != 7 {
		t.Fatalf("unexpected anytime count: %d", len(anytime))
	}
	found := map[string]bool{}
	for _, item := range anytime {
		found[item.Title] = true
	}
	if !found["Anytime Task"] || !found["Deadline Task"] || !found["Deadline Today"] || !found["Today Task"] || !found["Project Task"] || !found["Checklist Task"] || !found["Suppressed Deadline"] {
		t.Fatalf("missing anytime items: %+v", found)
	}
	if found["Recurring Task"] {
		t.Fatalf("unexpected anytime items: %+v", found)
	}

	someday, err := store.ListSomeday(0, 0)
	if err != nil {
		t.Fatalf("ListSomeday: %v", err)
	}
	if len(someday) != 1 || someday[0].Title != "Someday Task" {
		t.Fatalf("unexpected someday results: %+v", someday)
	}

	logbook, err := store.ListLogbook(0, 0)
	if err != nil {
		t.Fatalf("ListLogbook: %v", err)
	}
	if len(logbook) != 2 {
		t.Fatalf("unexpected logbook count: %d", len(logbook))
	}

	projects, err := store.ListProjects(0, 0)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 || projects[0].Title != "Project Alpha" {
		t.Fatalf("unexpected projects: %+v", projects)
	}

	filter := SearchFilter{Tag: "work"}
	work, err := store.Search(filter)
	if err != nil {
		t.Fatalf("Search tag: %v", err)
	}
	if len(work) != 1 || work[0].Title != "Project Task" {
		t.Fatalf("unexpected tag search: %+v", work)
	}

	filter = SearchFilter{Project: "Project Alpha"}
	byProject, err := store.Search(filter)
	if err != nil {
		t.Fatalf("Search project: %v", err)
	}
	if len(byProject) != 1 || byProject[0].Title != "Project Task" {
		t.Fatalf("unexpected project search: %+v", byProject)
	}

	filter = SearchFilter{Status: "done"}
	completed, err := store.Search(filter)
	if err != nil {
		t.Fatalf("Search status: %v", err)
	}
	if len(completed) != 1 || completed[0].Title != "Completed Task" {
		t.Fatalf("unexpected completed search: %+v", completed)
	}

	filter = SearchFilter{Deadline: "today"}
	deadlineToday, err := store.Search(filter)
	if err != nil {
		t.Fatalf("Search deadline: %v", err)
	}
	if len(deadlineToday) != 2 {
		t.Fatalf("unexpected deadline search: %+v", deadlineToday)
	}
	foundDeadline := map[string]bool{}
	for _, item := range deadlineToday {
		foundDeadline[item.Title] = true
	}
	if !foundDeadline["Deadline Today"] || !foundDeadline["Suppressed Deadline"] {
		t.Fatalf("unexpected deadline search: %+v", deadlineToday)
	}
}

func TestShowItemProjectAndChecklist(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	_, project, err := store.ShowItem("Project Alpha", true)
	if err != nil {
		t.Fatalf("ShowItem project: %v", err)
	}
	if project == nil || len(project.Todos) != 1 || len(project.Headings) != 1 {
		t.Fatalf("unexpected project details: %+v", project)
	}

	todo, _, err := store.ShowItem("Checklist Task", true)
	if err != nil {
		t.Fatalf("ShowItem checklist: %v", err)
	}
	if todo == nil || len(todo.Checklist) != 2 {
		t.Fatalf("unexpected checklist: %+v", todo)
	}
}

func setupTestStore(t *testing.T) (*Store, func()) {
	dsn := "file:thx-test?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	schema := []string{
		`CREATE TABLE TMTask (
			uuid TEXT PRIMARY KEY,
			title TEXT,
			status INTEGER,
			type INTEGER,
			start INTEGER,
			startDate INTEGER,
			deadline INTEGER,
			deadlineSuppressionDate INTEGER,
			trashed INTEGER,
			rt1_recurrenceRule TEXT,
			notes TEXT,
			project TEXT,
			area TEXT,
			heading TEXT,
			creationDate REAL,
			userModificationDate REAL,
			stopDate REAL,
			todayIndex INTEGER,
			"index" INTEGER
		)`,
		`CREATE TABLE TMTag (uuid TEXT PRIMARY KEY, title TEXT)`,
		`CREATE TABLE TMTaskTag (tasks TEXT, tags TEXT)`,
		`CREATE TABLE TMArea (uuid TEXT PRIMARY KEY, title TEXT)`,
		`CREATE TABLE TMChecklistItem (task TEXT, title TEXT, status INTEGER, "index" INTEGER)`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("schema exec: %v", err)
		}
	}

	now := time.Now()
	today := things.TimeToThingsDate(now)
	tomorrow := things.TimeToThingsDate(now.AddDate(0, 0, 1))
	future := things.TimeToThingsDate(now.AddDate(0, 0, 2))
	deadline := things.TimeToThingsDate(now.AddDate(0, 0, 3))

	// Areas
	if _, err := db.Exec("INSERT INTO TMArea (uuid, title) VALUES ('area-1', 'Personal')"); err != nil {
		t.Fatalf("insert area: %v", err)
	}

	// Tags
	if _, err := db.Exec("INSERT INTO TMTag (uuid, title) VALUES ('tag-1', 'work'), ('tag-2', 'errands')"); err != nil {
		t.Fatalf("insert tags: %v", err)
	}

	// Project
	if _, err := db.Exec("INSERT INTO TMTask (uuid, title, status, type, start, trashed, creationDate, userModificationDate, " +
		"area, \"index\") VALUES ('proj-1', 'Project Alpha', 0, 1, 1, 0, 1, 1, 'area-1', 1)"); err != nil {
		t.Fatalf("insert project: %v", err)
	}

	// Heading
	if _, err := db.Exec("INSERT INTO TMTask (uuid, title, status, type, start, trashed, project, creationDate, userModificationDate, \"index\") VALUES ('head-1', 'Heading A', 0, 2, 1, 0, 'proj-1', 1, 1, 1)"); err != nil {
		t.Fatalf("insert heading: %v", err)
	}

	// Tasks
	insertTask := func(id, title string, status, start, typ int, startDate, deadlineDate things.ThingsDate, project string, suppressionDate int64, recurrenceRule string) {
		_, err := db.Exec(`INSERT INTO TMTask (uuid, title, status, type, start, startDate, deadline, deadlineSuppressionDate, trashed, rt1_recurrenceRule, project, area, creationDate, userModificationDate, stopDate, todayIndex, "index")
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, 'area-1', 1, 1, 1, 1, 1)`,
			id, title, status, typ, start, nullDate(startDate), nullDate(deadlineDate), nullInt64(suppressionDate), nullStringValue(recurrenceRule), project)
		if err != nil {
			t.Fatalf("insert task %s: %v", id, err)
		}
	}

	insertTask("task-inbox", "Inbox Task", statusOpen, startInbox, typeTodo, 0, 0, "", 0, "")
	insertTask("task-today", "Today Task", statusOpen, startScheduled, typeTodo, today, 0, "", 0, "")
	insertTask("task-upcoming", "Upcoming Task", statusOpen, startSomeday, typeTodo, future, 0, "", 0, "")
	insertTask("task-deadline", "Deadline Task", statusOpen, startScheduled, typeTodo, 0, deadline, "", 0, "")
	insertTask("task-deadline-today", "Deadline Today", statusOpen, startScheduled, typeTodo, 0, today, "", 0, "")
	insertTask("task-anytime", "Anytime Task", statusOpen, startScheduled, typeTodo, 0, 0, "", 0, "")
	insertTask("task-someday", "Someday Task", statusOpen, startSomeday, typeTodo, 0, 0, "", 0, "")
	insertTask("task-completed", "Completed Task", statusDone, startScheduled, typeTodo, 0, 0, "", 0, "")
	insertTask("task-canceled", "Canceled Task", statusCanceled, startScheduled, typeTodo, 0, 0, "", 0, "")
	insertTask("task-project", "Project Task", statusOpen, startScheduled, typeTodo, tomorrow, 0, "proj-1", 0, "")
	insertTask("task-checklist", "Checklist Task", statusOpen, startScheduled, typeTodo, tomorrow, 0, "", 0, "")
	insertTask("task-suppressed-deadline", "Suppressed Deadline", statusOpen, startScheduled, typeTodo, 0, today, "", 1, "")
	insertTask("task-recurring", "Recurring Task", statusOpen, startScheduled, typeTodo, today, 0, "", 0, "FREQ=DAILY")

	// Tag linkage
	if _, err := db.Exec("INSERT INTO TMTaskTag (tasks, tags) VALUES ('task-project', 'tag-1')"); err != nil {
		t.Fatalf("insert task tag: %v", err)
	}

	// Checklist
	if _, err := db.Exec("INSERT INTO TMChecklistItem (task, title, status, " + "\"index\"" + ") VALUES ('task-checklist', 'Item A', 0, 1), ('task-checklist', 'Item B', 1, 2)"); err != nil {
		t.Fatalf("insert checklist: %v", err)
	}

	store := &Store{db: db, loc: time.Local}
	cleanup := func() {
		_ = db.Close()
	}
	return store, cleanup
}

func nullDate(date things.ThingsDate) any {
	if date == 0 {
		return nil
	}
	return int64(date)
}

func nullInt64(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullStringValue(value string) any {
	if value == "" {
		return nil
	}
	return value
}
