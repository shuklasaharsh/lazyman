package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// ManPage represents a manual page entry
type ManPage struct {
	Name        string
	Section     string
	Description string
	Path        string
}

// GetManPages retrieves all available man pages on the system
func GetManPages() ([]ManPage, error) {
	manPaths := getManPaths()
	pages := make(map[string]ManPage)

	for _, manPath := range manPaths {
		err := filepath.Walk(manPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip paths with errors
			}

			if info.IsDir() {
				return nil
			}

			// Check if it's a man page file (e.g., man1, man2, etc.)
			dir := filepath.Base(filepath.Dir(path))
			if strings.HasPrefix(dir, "man") && len(dir) > 3 {
				section := dir[3:]
				name := filepath.Base(path)

				// Remove common extensions
				name = strings.TrimSuffix(name, ".gz")
				name = strings.TrimSuffix(name, ".bz2")

				// Remove section suffix if present (e.g., ls.1 -> ls)
				parts := strings.Split(name, ".")
				if len(parts) > 1 {
					name = parts[0]
				}

				key := fmt.Sprintf("%s(%s)", name, section)
				if _, exists := pages[key]; !exists {
					pages[key] = ManPage{
						Name:    name,
						Section: section,
						Path:    path,
					}
				}
			}

			return nil
		})
		if err != nil {
			continue
		}
	}

	// Convert map to slice and sort
	result := make([]ManPage, 0, len(pages))
	for _, page := range pages {
		result = append(result, page)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// GetManContent retrieves the formatted content of a man page
func GetManContent(name, section string) (string, error) {
	var cmd *exec.Cmd
	if section != "" {
		cmd = exec.Command("man", section, name)
	} else {
		cmd = exec.Command("man", name)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get man page: %w", err)
	}

	return string(output), nil
}

// SearchManPages searches for man pages by keyword
func SearchManPages(query string) ([]ManPage, error) {
	// If query is "." or empty, return all pages
	if query == "." || query == "" {
		return GetManPages()
	}

	cmd := exec.Command("man", "-k", query)
	output, err := cmd.Output()
	if err != nil {
		// man -k returns exit status 1 when no results found, which is not really an error
		// Just return empty list
		return []ManPage{}, nil
	}

	lines := strings.Split(string(output), "\n")
	pages := make([]ManPage, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse "name (section) - description" format
		parts := strings.SplitN(line, "-", 2)
		if len(parts) < 2 {
			continue
		}

		nameSection := strings.TrimSpace(parts[0])
		description := strings.TrimSpace(parts[1])

		// Extract name and section
		if idx := strings.Index(nameSection, "("); idx != -1 {
			name := strings.TrimSpace(nameSection[:idx])
			section := strings.Trim(nameSection[idx:], "()")

			pages = append(pages, ManPage{
				Name:        name,
				Section:     section,
				Description: description,
			})
		}
	}

	return pages, nil
}

// getManPaths returns common man page directories
func getManPaths() []string {
	paths := []string{
		"/usr/share/man",
		"/usr/local/share/man",
		"/opt/homebrew/share/man",
	}

	// Also check MANPATH environment variable
	if manpath := os.Getenv("MANPATH"); manpath != "" {
		for _, p := range strings.Split(manpath, ":") {
			if p != "" {
				paths = append(paths, p)
			}
		}
	}

	return paths
}

// GetRawManContent reads the raw man page file directly (much faster than calling man command)
func GetRawManContent(path string) (string, error) {
	// Check if file is gzipped
	if strings.HasSuffix(path, ".gz") {
		return readGzippedFile(path)
	}

	// Read regular file
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// readGzippedFile reads and decompresses a gzipped file
func readGzippedFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gz); err != nil {
		return "", err
	}

	return buf.String(), nil
}
