package skills

import (
	"context"
	"fmt"
)

// SkillsTool provides skill discovery and loading functionality
type SkillsTool struct {
	SkillsDirectory string
}

// NewSkillsTool creates a new SkillsTool
func NewSkillsTool(skillsDirectory string) *SkillsTool {
	return &SkillsTool{SkillsDirectory: skillsDirectory}
}

// Execute executes the skills tool command
func (t *SkillsTool) Execute(ctx context.Context, command string) (string, error) {
	if command == "" {
		// Return list of available skills
		discoveredSkills, err := DiscoverSkills(t.SkillsDirectory)
		if err != nil {
			return "", fmt.Errorf("failed to discover skills: %w", err)
		}
		return GenerateSkillsToolDescription(discoveredSkills), nil
	}

	// Load specific skill content
	content, err := LoadSkillContent(t.SkillsDirectory, command)
	if err != nil {
		return "", err
	}
	return content, nil
}

// BashTool provides shell command execution in skills context
type BashTool struct {
	SkillsDirectory string
}

// NewBashTool creates a new BashTool
func NewBashTool(skillsDirectory string) *BashTool {
	return &BashTool{SkillsDirectory: skillsDirectory}
}

// Execute executes a bash command in the skills context
func (t *BashTool) Execute(ctx context.Context, command string, sessionID string) (string, error) {
	// Get session path for working directory
	sessionPath, err := GetSessionPath(sessionID, t.SkillsDirectory)
	if err != nil {
		return "", fmt.Errorf("failed to get session path: %w", err)
	}

	return ExecuteCommand(ctx, command, sessionPath)
}

// FileTools provides file operation tools with path boundary enforcement.
// SessionPath is the working directory for the current session.
// SkillsDirectory is the read-only skills directory (only for read operations).
type FileTools struct {
	SessionPath     string
	SkillsDirectory string
}

// NewFileTools creates a new FileTools with path boundary enforcement.
func NewFileTools(sessionPath, skillsDirectory string) *FileTools {
	return &FileTools{
		SessionPath:     sessionPath,
		SkillsDirectory: skillsDirectory,
	}
}

// ReadFile reads a file with line numbers.
// Allows access to both session path and skills directory (read-only access to skills).
func (ft *FileTools) ReadFile(path string, offset, limit int) (string, error) {
	allowedRoots := []string{ft.SessionPath, ft.SkillsDirectory}
	return ReadFileContent(path, offset, limit, allowedRoots)
}

// WriteFile writes content to a file.
// Only allows access to session path (no write access to skills directory).
func (ft *FileTools) WriteFile(path string, content string) error {
	allowedRoots := []string{ft.SessionPath}
	return WriteFileContent(path, content, allowedRoots)
}

// EditFile performs an exact string replacement in a file.
// Only allows access to session path (no write access to skills directory).
func (ft *FileTools) EditFile(path string, oldString, newString string, replaceAll bool) error {
	allowedRoots := []string{ft.SessionPath}
	return EditFileContent(path, oldString, newString, replaceAll, allowedRoots)
}

// InitializeSessionPath initializes a session's working directory with skills symlink
func InitializeSessionPath(sessionID, skillsDirectory string) (string, error) {
	return GetSessionPath(sessionID, skillsDirectory)
}
