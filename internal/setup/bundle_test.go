package setup

import (
	"io/fs"
	"reflect"
	"slices"
	"testing"
)

func TestListBundledSkillsExposesOnlyPublicCatalog(t *testing.T) {
	t.Parallel()

	skills, err := ListBundledSkills()
	if err != nil {
		t.Fatalf("list bundled skills: %v", err)
	}

	var names []string
	for _, skill := range skills {
		names = append(names, skill.Name)
	}

	want := []string{
		"rc",
		"rc-analyze",
		"rc-code-review",
		"rc-create-prd",
		"rc-create-tasks",
		"rc-create-techspec",
		"rc-execute-task",
		"rc-final-verify",
		"rc-fix-analysis",
		"rc-fix-reviews",
		"rc-git",
		"rc-jira",
		"rc-new-project",
		"rc-openapi",
		"rc-postman",
		"rc-project-memory",
		"rc-readme",
		"rc-review-round",
		"rc-simplify-review",
		"rc-workflow-memory",
	}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("unexpected bundled skill names\nwant: %#v\ngot:  %#v", want, names)
	}

	for _, forbidden := range []string{"brainstorming", "golang-pro", "testing-anti-patterns"} {
		if slices.Contains(names, forbidden) {
			t.Fatalf("expected internal skill %q to be excluded from bundled catalog", forbidden)
		}
	}
}

func TestBundledWorkflowMemorySkillIncludesReferenceFile(t *testing.T) {
	t.Parallel()

	bundle, err := bundledSkillsRoot()
	if err != nil {
		t.Fatalf("bundled skills root: %v", err)
	}

	if _, err := fs.Stat(bundle, "rc-workflow-memory/SKILL.md"); err != nil {
		t.Fatalf("expected bundled workflow-memory skill, got %v", err)
	}
	if _, err := fs.Stat(bundle, "rc-workflow-memory/references/memory-guidelines.md"); err != nil {
		t.Fatalf("expected bundled workflow-memory reference file, got %v", err)
	}
}

func TestListBundledReusableAgentsAllowsEmptyRoster(t *testing.T) {
	t.Run("Should return empty bundled reusable-agent roster when none exist", func(t *testing.T) {
		t.Parallel()

		reusableAgents, err := ListBundledReusableAgents()
		if err != nil {
			t.Fatalf("list bundled reusable agents: %v", err)
		}
		if len(reusableAgents) != 0 {
			t.Fatalf("expected bundled reusable-agent roster to be empty, got %#v", reusableAgents)
		}
	})
}

func TestBundledReusableAgentsRootRemainsReadableWhenEmpty(t *testing.T) {
	t.Run("Should keep bundled reusable agents root readable when the roster is empty", func(t *testing.T) {
		t.Parallel()

		bundle, err := bundledReusableAgentsRoot()
		if err != nil {
			t.Fatalf("bundled reusable agents root: %v", err)
		}
		if _, err := fs.ReadDir(bundle, "."); err != nil {
			t.Fatalf("ReadDir(.) error = %v", err)
		}
	})
}
