package tools

import (
	"fmt"

	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	"github.com/kagent-dev/kagent/go/adk/pkg/skills"
)

// bashTool executes shell commands safely in the skills environment.
// This tool uses the Anthropic Sandbox Runtime (srt) to execute commands with:
// - Filesystem restrictions (controlled read/write access)
// - Network restrictions (controlled domain access)
// - Process isolation at the OS level
type bashTool struct {
	skillsDirectory string
}

// NewBashTool creates a new bash tool backed by the given skills directory.
func NewBashTool(skillsDir string) tool.Tool {
	return &bashTool{skillsDirectory: skillsDir}
}

func (t *bashTool) Name() string {
	return "bash"
}

func (t *bashTool) Description() string {
	return "Execute bash commands safely in the skills environment. Use this tool for command-line operations like running scripts, installing packages, or managing files. For file operations (read/write/edit), use the dedicated file tools instead."
}

func (t *bashTool) IsLongRunning() bool {
	return false
}

// Declaration returns the function declaration for the LLM.
func (t *bashTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: "OBJECT",
			Properties: map[string]*genai.Schema{
				"command": {
					Type:        "STRING",
					Description: "Bash command to execute. Use && to chain commands.",
				},
				"description": {
					Type:        "STRING",
					Description: "Clear, concise description of what this command does (5-10 words)",
				},
			},
			Required: []string{"command"},
		},
	}
}

// Run executes the bash command using the skills package.
func (t *bashTool) Run(toolCtx tool.Context, args any) (map[string]any, error) {
	m, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected args type, got: %T", args)
	}

	commandRaw, exists := m["command"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: command")
	}

	command, ok := commandRaw.(string)
	if !ok {
		return nil, fmt.Errorf("command must be a string, got: %T", commandRaw)
	}

	// Get session path for working directory
	sessionPath, err := skills.GetSessionPath(toolCtx.SessionID(), t.skillsDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to get session path: %w", err)
	}

	// Execute the command
	result, err := skills.ExecuteCommand(toolCtx, command, sessionPath)
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	return map[string]any{
		"output": result,
	}, nil
}

// ProcessRequest packs the tool's function declaration into the LLM request.
func (t *bashTool) ProcessRequest(_ tool.Context, req *model.LLMRequest) error {
	if req.Tools == nil {
		req.Tools = make(map[string]any)
	}
	if _, ok := req.Tools[t.Name()]; ok {
		return fmt.Errorf("duplicate tool: %q", t.Name())
	}
	req.Tools[t.Name()] = t

	if req.Config == nil {
		req.Config = &genai.GenerateContentConfig{}
	}
	// Find an existing genai.Tool with FunctionDeclarations or create one.
	var funcTool *genai.Tool
	for _, gt := range req.Config.Tools {
		if gt != nil && gt.FunctionDeclarations != nil {
			funcTool = gt
			break
		}
	}
	if funcTool == nil {
		req.Config.Tools = append(req.Config.Tools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{t.Declaration()},
		})
	} else {
		funcTool.FunctionDeclarations = append(funcTool.FunctionDeclarations, t.Declaration())
	}
	return nil
}
