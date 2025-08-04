package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  check-labels <labels-json>     - Check for release labels\n")
		fmt.Fprintf(os.Stderr, "  bump-version <release-type>    - Bump version in VERSION file\n")
		fmt.Fprintf(os.Stderr, "  create-notes <pr-data-json>    - Create release notes\n")
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
	// Read current version from project root
	// Try different paths since we might be in tools/release-manager or project root
	versionPaths := []string{"../../VERSION", "VERSION"}
	var versionBytes []byte
	var err error
	
	for _, path := range versionPaths {
		versionBytes, err = os.ReadFile(path)
		if err == nil {
			fmt.Printf("Found VERSION file at: %s\n", path)
			break
		}
	}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading VERSION file from any path: %v\n", err)
		os.Exit(1)
	}

	currentVersion := strings.TrimSpace(string(versionBytes))
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

	// Write new version to project root
	// Use the same path logic for writing
	writePaths := []string{"../../VERSION", "VERSION"}
	var writeErr error
	
	for _, path := range writePaths {
		writeErr = os.WriteFile(path, []byte(newVersion), 0644)
		if writeErr == nil {
			fmt.Printf("Updated VERSION file at: %s\n", path)
			break
		}
	}
	
	if writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing VERSION file to any path: %v\n", writeErr)
		os.Exit(1)
	}

	// Output for GitHub Actions
	fmt.Printf("old-version=%s\n", currentVersion)
	fmt.Printf("new-version=%s\n", newVersion)
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
