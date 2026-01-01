package things

import "testing"

func TestBuildAddURL(t *testing.T) {
	url := BuildAddURL(AddOptions{
		Title:     "Buy milk",
		When:      "today",
		Deadline:  "2024-01-10",
		Tags:      []string{"urgent", "errands"},
		Project:   "Groceries",
		Notes:     "From store",
		Checklist: []string{"Eggs", "Milk"},
	})
	want := "things:///add?checklist-items=Eggs%0AMilk&deadline=2024-01-10&list=Groceries&notes=From%20store&tags=errands%2Curgent&title=Buy%20milk&when=today"
	if url != want {
		t.Fatalf("unexpected url:\nwant: %s\n got: %s", want, url)
	}
}

func TestBuildUpdateURL(t *testing.T) {
	url := BuildUpdateURL(UpdateOptions{
		ID:       "ABC123",
		Title:    "New title",
		When:     "tomorrow",
		Deadline: "2024-02-01",
		Tags:     []string{"work"},
		Notes:    "Note",
	})
	want := "things:///update?deadline=2024-02-01&id=ABC123&notes=Note&tags=work&title=New%20title&when=tomorrow"
	if url != want {
		t.Fatalf("unexpected url:\nwant: %s\n got: %s", want, url)
	}
}
