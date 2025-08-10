package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
    if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  check-labels <labels-json>     - Check for release labels\n")
        fmt.Fprintf(os.Stderr, "  bump-version <release-type>    - Compute next version from latest git tag\n")
		fmt.Fprintf(os.Stderr, "  create-notes <pr-data-json>    - Create release notes\n")
		fmt.Fprintf(os.Stderr, "  create-tag <new-version> <old-version> <pr-number> <release-type> - Create and push a git tag\n")
		fmt.Fprintf(os.Stderr, "  publish-release <version> <release-type>   - Create GitHub Release with artifacts\n")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "check-labels":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s check-labels <labels-json>\n", os.Args[0])
			os.Exit(1)
		}
		checkLabels(os.Args[2])
	case "bump-version":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s bump-version <release-type>\n", os.Args[0])
			os.Exit(1)
		}
		bumpVersion(os.Args[2])
	case "create-notes":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s create-notes <pr-data-json>\n", os.Args[0])
			os.Exit(1)
		}
		createReleaseNotes(os.Args[2])
	case "create-tag":
		if len(os.Args) != 6 {
			fmt.Fprintf(os.Stderr, "Usage: %s create-tag <new-version> <old-version> <pr-number> <release-type>\n", os.Args[0])
			os.Exit(1)
		}
		createTag(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	case "publish-release":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "Usage: %s publish-release <version> <release-type>\n", os.Args[0])
			os.Exit(1)
		}
		publishRelease(os.Args[2], os.Args[3])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func checkLabels(labelsJSON string) {
	var labels []string
	if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing labels JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PR Labels: %v\n", labels)

	// Check for release labels in priority order
	releaseTypes := []string{"major", "minor", "patch", "alpha", "beta", "rc"}

	for _, releaseType := range releaseTypes {
		labelName := "release:" + releaseType
		for _, label := range labels {
			if label == labelName {
				fmt.Printf("release-type=%s\n", releaseType)
				fmt.Printf("🎯 %s release detected\n", strings.Title(releaseType))
				return
			}
		}
	}

	fmt.Printf("release-type=none\n")
	fmt.Printf("ℹ️ No release labels found - skipping release\n")
}

func bumpVersion(releaseType string) {
    currentVersion, err := getLatestVersionFromGitTag()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error determining current version from git tags: %v\n", err)
        os.Exit(1)
    }
	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Release type: %s\n", releaseType)

	// Parse version using regex
	versionRegex := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(-([a-zA-Z]+)\.(\d+))?$`)
	matches := versionRegex.FindStringSubmatch(currentVersion)

	if matches == nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid version format: %s\n", currentVersion)
		os.Exit(1)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	var prereleaseName string
	var prereleaseNum int
	if matches[4] != "" { // Has prerelease
		prereleaseName = matches[5]
		prereleaseNum, _ = strconv.Atoi(matches[6])
	}

	var newVersion string

	switch releaseType {
	case "major":
		newVersion = fmt.Sprintf("%d.0.0", major+1)
	case "minor":
		newVersion = fmt.Sprintf("%d.%d.0", major, minor+1)
	case "patch":
		newVersion = fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
	case "alpha", "beta", "rc":
		if prereleaseName == releaseType {
			// Same prerelease type, bump number
			newVersion = fmt.Sprintf("%d.%d.%d-%s.%d", major, minor, patch, releaseType, prereleaseNum+1)
		} else if prereleaseName != "" {
			// Different prerelease type
			newVersion = fmt.Sprintf("%d.%d.%d-%s.1", major, minor, patch, releaseType)
		} else {
			// Not a prerelease, bump patch and add prerelease
			newVersion = fmt.Sprintf("%d.%d.%d-%s.1", major, minor, patch+1, releaseType)
		}
	default:
		fmt.Fprintf(os.Stderr, "❌ Invalid release type: %s\n", releaseType)
		os.Exit(1)
	}

	fmt.Printf("New version: %s\n", newVersion)

	// Output for GitHub Actions
	fmt.Printf("old-version=%s\n", currentVersion)
	fmt.Printf("new-version=%s\n", newVersion)
}

func getLatestVersionFromGitTag() (string, error) {
    // Ensure we are at repo root or any directory within repo. We will call git to get the latest version tag.
    // Strategy: use `git tag --list 'v*' --sort=-v:refname | head -n1`
    cmd := exec.Command("git", "tag", "--list", "v*", "--sort=-v:refname")
    output, err := cmd.Output()
    if err != nil {
        // Fallback: try describe
        describeCmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
        out2, err2 := describeCmd.Output()
        if err2 != nil {
            // No tags present; start from 0.0.0
            return "0.0.0", nil
        }
        tag := strings.TrimSpace(string(out2))
        tag = strings.TrimPrefix(tag, "v")
        return tag, nil
    }
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
        // No tags; start at 0.0.0
        return "0.0.0", nil
    }
    latestTag := strings.TrimSpace(lines[0])
    latestTag = strings.TrimPrefix(latestTag, "v")
    return latestTag, nil
}

type PRData struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
	Type   string `json:"release_type"`
}

func createReleaseNotes(prDataJSON string) {
	var prData PRData
	if err := json.Unmarshal([]byte(prDataJSON), &prData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing PR data JSON: %v\n", err)
		os.Exit(1)
	}

	releaseNotes := fmt.Sprintf(`## 🎉 Auto-Release from PR #%d

**Release Type:** %s  
**Triggered by:** @%s

### 📋 Changes from PR #%d

**%s**

%s

---

*This release was automatically created from the merged pull request.*
`, prData.Number, prData.Type, prData.Author, prData.Number, prData.Title, prData.Body)

	if err := os.WriteFile("release-notes.md", []byte(releaseNotes), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing release notes: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📝 Release notes created in release-notes.md\n")
}

func createTag(newVersion, oldVersion, prNumber, releaseType string) {
	_ = oldVersion

	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+([-][A-Za-z]+\.[0-9]+)?$`)
	if !versionRegex.MatchString(newVersion) {
		fmt.Fprintf(os.Stderr, "❌ Invalid version format: %s\n", newVersion)
		os.Exit(1)
	}

	fmt.Printf("Creating release tag v%s\n", newVersion)

	_ = exec.Command("git", "config", "--local", "user.email", "action@github.com").Run()
	_ = exec.Command("git", "config", "--local", "user.name", "GitHub Action").Run()

	if err := exec.Command("git", "rev-parse", "v"+newVersion).Run(); err == nil {
		fmt.Printf("⚠️ Tag v%s already exists, nothing to do\n", newVersion)
		return
	}

	tagMsg := fmt.Sprintf("Release %s (from PR #%s, %s)", newVersion, prNumber, releaseType)
	if err := exec.Command("git", "tag", "-m", tagMsg, "v"+newVersion).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tag: %v\n", err)
		os.Exit(1)
	}
	if err := exec.Command("git", "push", "origin", "v"+newVersion).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error pushing tag: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🎉 Created and pushed tag v%s\n", newVersion)
	fmt.Printf("🔗 Tag: v%s\n", newVersion)
	fmt.Printf("📋 Triggered by PR #%s (%s release)\n", prNumber, releaseType)
}

func publishRelease(version, releaseType string) {
	fmt.Printf("🚀 Creating GitHub Release...\n")

	prerelease := releaseType == "alpha" || releaseType == "beta" || releaseType == "rc"

	if err := os.MkdirAll("dist", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dist directory: %v\n", err)
		os.Exit(1)
	}

	readme := fmt.Sprintf(`# Vex Language v%[1]s

Basic transpiler with working features:
- Variable definitions: (def x 10) → x := 10
- Arithmetic expressions: (+ 1 2) → 1 + 2
- CLI tool: vex-transpiler
`, version)
	if err := os.WriteFile("dist/README.md", []byte(readme), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing dist/README.md: %v\n", err)
		os.Exit(1)
	}

	archive := fmt.Sprintf("dist/vex-examples-v%[1]s.tar.gz", version)
	if err := exec.Command("tar", "-czf", archive, "examples/").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating examples archive: %v\n", err)
		os.Exit(1)
	}

	notesPath := os.Getenv("RELEASE_NOTES_PATH")
	if notesPath == "" {
		notesPath = "/tmp/release-notes.md"
	}
	notesBytes, err := os.ReadFile(notesPath)
	if err != nil {
		// Fallback to repo-local file created by create-notes
		if alt, err2 := os.ReadFile("release-notes.md"); err2 == nil {
			notesBytes = alt
		} else {
			notesBytes = []byte(fmt.Sprintf("Vex v%s", version))
		}
	}
	notes := string(notesBytes)

	args := []string{"release", "create", "v" + version, "--title", "Vex v" + version, "--notes", notes}
	if prerelease {
		args = append(args, "--prerelease")
	}

	entries, err := os.ReadDir("dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading dist directory: %v\n", err)
		os.Exit(1)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		args = append(args, filepath.Join("dist", e.Name()))
	}

	cmd := exec.Command("gh", args...)
	out, err := cmd.CombinedOutput()
	fmt.Print(string(out))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating GitHub release: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ GitHub Release v%s created successfully\n", version)
}
