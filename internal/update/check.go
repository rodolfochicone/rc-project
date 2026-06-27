package update

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	selfupdate "github.com/creativeprojects/go-selfupdate"
)

const (
	checkInterval        = 24 * time.Hour
	noUpdateNotifierEnv  = "RC_NO_UPDATE_NOTIFIER"
	repositorySlugString = "rodolfochicone/rc-project"
)

var (
	nowFunc          = time.Now
	newUpdaterClient = newSelfUpdaterClient
)

// ReleaseInfo holds metadata about an available release.
type ReleaseInfo struct {
	Version     string    `yaml:"version"`
	URL         string    `yaml:"url"`
	PublishedAt time.Time `yaml:"published_at"`
}

type updaterClient interface {
	DetectLatest(context.Context) (*ReleaseInfo, error)
	UpdateSelf(context.Context, string) (*ReleaseInfo, error)
}

type selfUpdaterClient struct {
	updater *selfupdate.Updater
}

// ShouldCheckForUpdate reports whether the cached state is stale enough to justify
// a network request for a newer release.
func ShouldCheckForUpdate(state *StateEntry, now time.Time) bool {
	if state == nil {
		return true
	}
	if state.CheckedForUpdateAt.IsZero() {
		return true
	}
	return now.Sub(state.CheckedForUpdateAt) >= checkInterval
}

// CheckForUpdate returns release metadata when a newer version is available.
//
// The function skips all work for dev builds and when RC_NO_UPDATE_NOTIFIER is set.
// Cached results are reused for 24 hours to avoid redundant GitHub API calls.
func CheckForUpdate(ctx context.Context, currentVersion, stateFilePath string) (*ReleaseInfo, error) {
	if strings.TrimSpace(currentVersion) == "dev" {
		slog.Debug("skip update check for dev build")
		return nil, nil
	}
	if _, ok := os.LookupEnv(noUpdateNotifierEnv); ok {
		slog.Debug("skip update check because notifier is disabled", slog.String("env", noUpdateNotifierEnv))
		return nil, nil
	}
	if githubToken() == "" {
		slog.Debug("skip update check because no GitHub token is available")
		return nil, nil
	}

	state, err := ReadState(stateFilePath)
	if err != nil {
		return nil, err
	}

	now := nowFunc().UTC()
	if !ShouldCheckForUpdate(state, now) {
		slog.Debug("reuse cached update state", slog.Time("checked_at", state.CheckedForUpdateAt))
		return newerRelease(currentVersion, releaseInfoPtr(state.LatestRelease))
	}

	slog.Debug("checking for a newer rc release")
	client, err := newUpdaterClient()
	if err != nil {
		return nil, err
	}

	latest, err := client.DetectLatest(ctx)
	if err != nil {
		return nil, err
	}

	entry := &StateEntry{CheckedForUpdateAt: now}
	if latest != nil {
		entry.LatestRelease = *latest
	}
	if err := WriteState(stateFilePath, entry); err != nil {
		return nil, err
	}

	return newerRelease(currentVersion, latest)
}

// githubToken resolves the GitHub API token used to reach the release
// repository, checking GITHUB_TOKEN first and falling back to GH_TOKEN.
//
// While the release repository is private, the passive update notifier is
// skipped entirely when no token is set: an unauthenticated API request would
// fail and add latency to every command. Revisit this guard if the repository
// becomes public, where unauthenticated checks succeed.
func githubToken() string {
	for _, key := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func newSelfUpdaterClient() (updaterClient, error) {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{APIToken: githubToken()})
	if err != nil {
		return nil, fmt.Errorf("create update source: %w", err)
	}
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return nil, fmt.Errorf("create self-update client: %w", err)
	}
	return &selfUpdaterClient{updater: updater}, nil
}

func (c *selfUpdaterClient) DetectLatest(ctx context.Context) (*ReleaseInfo, error) {
	release, found, err := c.updater.DetectLatest(ctx, selfupdate.ParseSlug(repositorySlugString))
	if err != nil {
		return nil, fmt.Errorf("detect latest release: %w", err)
	}
	if !found || release == nil {
		return nil, nil
	}

	info := releaseInfoFromSelfUpdate(release)
	slog.Debug("found latest rc release", slog.String("latest_version", info.Version))
	return info, nil
}

func (c *selfUpdaterClient) UpdateSelf(ctx context.Context, currentVersion string) (*ReleaseInfo, error) {
	release, err := c.updater.UpdateSelf(ctx, currentVersion, selfupdate.ParseSlug(repositorySlugString))
	if err != nil {
		return nil, fmt.Errorf("self-update rc: %w", err)
	}
	if release == nil {
		return nil, nil
	}

	return releaseInfoFromSelfUpdate(release), nil
}

func releaseInfoFromSelfUpdate(release *selfupdate.Release) *ReleaseInfo {
	if release == nil {
		return nil
	}
	version := strings.TrimSpace(release.Version())
	if version == "" {
		return nil
	}

	return &ReleaseInfo{
		Version:     version,
		URL:         release.URL,
		PublishedAt: release.PublishedAt,
	}
}

// resolveCurrentVersionForUpdate maps the running binary's version to a value the
// self-update flow can compare. A dev build (or any value that is not a valid
// release version) becomes "0.0.0" so every published release is treated as newer
// and the upgrade proceeds, instead of failing semantic-version parsing.
func resolveCurrentVersionForUpdate(currentVersion string) string {
	if _, err := parseVersion(currentVersion); err != nil {
		return "0.0.0"
	}
	return currentVersion
}

func newerRelease(currentVersion string, latest *ReleaseInfo) (*ReleaseInfo, error) {
	if latest == nil || strings.TrimSpace(latest.Version) == "" {
		return nil, nil
	}

	current, err := parseVersion(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("parse current version %q: %w", currentVersion, err)
	}
	release, err := parseVersion(latest.Version)
	if err != nil {
		return nil, fmt.Errorf("parse latest version %q: %w", latest.Version, err)
	}

	if current.Compare(release) >= 0 {
		return nil, nil
	}
	return latest, nil
}

func parseVersion(raw string) (*semver.Version, error) {
	return semver.NewVersion(strings.TrimPrefix(trimGitDescribeSuffix(raw), "v"))
}

func trimGitDescribeSuffix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	commitSep := strings.LastIndex(trimmed, "-g")
	if commitSep < 0 || commitSep+2 >= len(trimmed) {
		return trimmed
	}
	commit := trimmed[commitSep+2:]
	if !isGitShortSHA(commit) {
		return trimmed
	}
	beforeCommit := trimmed[:commitSep]
	countSep := strings.LastIndex(beforeCommit, "-")
	if countSep < 0 || countSep+1 >= len(beforeCommit) {
		return trimmed
	}
	for _, char := range beforeCommit[countSep+1:] {
		if char < '0' || char > '9' {
			return trimmed
		}
	}
	return beforeCommit[:countSep]
}

func isGitShortSHA(value string) bool {
	if len(value) < 7 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}

func releaseInfoPtr(info ReleaseInfo) *ReleaseInfo {
	if strings.TrimSpace(info.Version) == "" && strings.TrimSpace(info.URL) == "" && info.PublishedAt.IsZero() {
		return nil
	}
	cloned := info
	return &cloned
}
