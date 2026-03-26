package tools

import (
	"google.golang.org/adk/tool"
)

// AddSkillsToolsToAgent adds all skills-related tools to the agent's tool list.
// If skillsDir is empty, no tools are added.
//
// This function adds the following tools:
// - skills: Discover and load specialized domain skills
// - bash: Execute shell commands safely
// - read_file: Read files with line numbers
// - write_file: Write content to files
// - edit_file: Edit files with exact string replacement
func AddSkillsToolsToAgent(skillsDir string, extraTools *[]tool.Tool) {
	if skillsDir == "" {
		return
	}
	*extraTools = append(*extraTools,
		NewSkillsTool(skillsDir),
		NewBashTool(skillsDir),
		NewReadFileTool(skillsDir),
		NewWriteFileTool(skillsDir),
		NewEditFileTool(skillsDir),
	)
}
