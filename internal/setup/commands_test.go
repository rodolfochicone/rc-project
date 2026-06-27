package setup

import (
	"strings"
	"testing"
)

func TestListBundledCommands(t *testing.T) {
	t.Parallel()

	commands, err := ListBundledCommands()
	if err != nil {
		t.Fatalf("list bundled commands: %v", err)
	}

	got := make([]string, 0, len(commands))
	for i := range commands {
		got = append(got, commands[i].Name)
	}
	want := []string{"rc-docs", "rc-exec", "rc-pipe", "rc-plan", "rc-review"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("bundled commands = %v, want %v", got, want)
	}

	for i := range commands {
		if strings.TrimSpace(commands[i].Description) == "" {
			t.Errorf("command %q has empty description", commands[i].Name)
		}
		if commands[i].FileName != commands[i].Name+".md" {
			t.Errorf("command %q FileName = %q, want %s.md", commands[i].Name, commands[i].FileName, commands[i].Name)
		}
	}
}

func TestListCommandsParsesFlatMarkdownAndIgnoresNoise(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, map[string]string{
		"rc-plan.md": "---\ndescription: Plan a feature\n---\nbody\n",
		"rc-pipe.md": "---\ndescription: Full pipeline\n---\nbody\n",
		"README.md":  "---\ndescription: should be ignored\n---\n",
		"notes.txt":  "not markdown\n",
	})

	commands, err := ListCommands(bundle)
	if err != nil {
		t.Fatalf("list commands: %v", err)
	}

	got := make([]string, 0, len(commands))
	for i := range commands {
		got = append(got, commands[i].Name)
	}
	want := []string{"rc-pipe", "rc-plan"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("commands = %v, want %v (README.md and non-.md files must be ignored)", got, want)
	}
}

func TestListCommandsRejectsMissingDescription(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, map[string]string{
		"rc-plan.md": "---\nargument-hint: [feature]\n---\nbody\n",
	})

	_, err := ListCommands(bundle)
	if err == nil || !strings.Contains(err.Error(), "missing description") {
		t.Fatalf("expected missing description error, got %v", err)
	}
}

func TestListCommandsRejectsInvalidFrontmatter(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, map[string]string{
		"rc-plan.md": "---\ndescription: [a] [b]\n---\nbody\n",
	})

	_, err := ListCommands(bundle)
	if err == nil {
		t.Fatal("expected invalid frontmatter to fail")
	}
}

func TestSelectCommands(t *testing.T) {
	t.Parallel()

	all := []Command{{Name: "rc-plan"}, {Name: "rc-pipe"}, {Name: "rc-docs"}}

	t.Run("selects requested commands by name", func(t *testing.T) {
		t.Parallel()

		selected, err := SelectCommands(all, []string{"rc-pipe", "rc-plan"})
		if err != nil {
			t.Fatalf("select commands: %v", err)
		}
		if len(selected) != 2 {
			t.Fatalf("selected %d commands, want 2", len(selected))
		}
	})

	t.Run("rejects an unknown command name", func(t *testing.T) {
		t.Parallel()

		if _, err := SelectCommands(all, []string{"nope"}); err == nil {
			t.Fatal("expected unknown command error")
		}
	})
}
