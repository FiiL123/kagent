package tools

import (
	"fmt"

	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	"github.com/kagent-dev/kagent/go/adk/pkg/skills"
)

// skillsTool provides skill discovery and loading functionality.
// Skills are specialized domain knowledge and scripts that the agent can use
// to solve complex tasks.
type skillsTool struct {
	skillsDirectory string
}

// NewSkillsTool creates a new skills tool backed by the given skills directory.
func NewSkillsTool(skillsDir string) tool.Tool {
	return &skillsTool{skillsDirectory: skillsDir}
}

func (t *skillsTool) Name() string {
	return "skills"
}

func (t *skillsTool) Description() string {
	return "Discover and load specialized domain skills. Skills provide domain-specific knowledge, tools, and instructions that extend agent capabilities. Use this to find available skills or load specific skill instructions when needed."
}

func (t *skillsTool) IsLongRunning() bool {
	return false
}

// Declaration returns the function declaration for the LLM.
func (t *skillsTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: "OBJECT",
			Properties: map[string]*genai.Schema{
				"command": {
					Type:        "STRING",
					Description: "Optional skill name to load. If empty, returns list of available skills.",
				},
			},
		},
	}
}

// Run executes the skills tool command.
func (t *skillsTool) Run(toolCtx tool.Context, args any) (map[string]any, error) {
	m, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected args type, got: %T", args)
	}

	var command string
	if commandRaw, exists := m["command"]; exists {
		command, ok = commandRaw.(string)
		if !ok {
			return nil, fmt.Errorf("command must be a string, got: %T", commandRaw)
		}
	}

	// Create a new SkillsTool and execute
	st := skills.NewSkillsTool(t.skillsDirectory)

	result, err := st.Execute(toolCtx, command)
	if err != nil {
		return nil, fmt.Errorf("skills execution failed: %w", err)
	}

	return map[string]any{
		"output": result,
	}, nil
}

// ProcessRequest packs the tool's function declaration into the LLM request.
func (t *skillsTool) ProcessRequest(_ tool.Context, req *model.LLMRequest) error {
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
