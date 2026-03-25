package tools

import (
	"fmt"

	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	"github.com/kagent-dev/kagent/go/adk/pkg/skills"
)

// readFileTool reads a file with line numbers.
type readFileTool struct {
	skillsDirectory string
}

// NewReadFileTool creates a new read file tool.
func NewReadFileTool(skillsDir string) tool.Tool {
	return &readFileTool{skillsDirectory: skillsDir}
}

func (t *readFileTool) Name() string {
	return "read_file"
}

func (t *readFileTool) Description() string {
	return "Read a file's contents with line numbers. Use this to view file contents when you need to understand or analyze code, configuration files, or text documents."
}

func (t *readFileTool) IsLongRunning() bool {
	return false
}

func (t *readFileTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: "OBJECT",
			Properties: map[string]*genai.Schema{
				"path": {
					Type:        "STRING",
					Description: "Path to the file to read",
				},
				"offset": {
					Type:        "INTEGER",
					Description: "Optional line number to start reading from (1-based, default: 1)",
				},
				"limit": {
					Type:        "INTEGER",
					Description: "Optional maximum number of lines to read (default: all)",
				},
			},
			Required: []string{"path"},
		},
	}
}

func (t *readFileTool) Run(toolCtx tool.Context, args any) (map[string]any, error) {
	m, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected args type, got: %T", args)
	}

	pathRaw, exists := m["path"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: path")
	}

	path, ok := pathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string, got: %T", pathRaw)
	}

	// Extract optional offset parameter
	var offset int
	if offsetRaw, exists := m["offset"]; exists {
		switch v := offsetRaw.(type) {
		case int64:
			offset = int(v)
		case float64:
			offset = int(v)
		case int:
			offset = v
		}
	}

	// Extract optional limit parameter
	var limit int
	if limitRaw, exists := m["limit"]; exists {
		switch v := limitRaw.(type) {
		case int64:
			limit = int(v)
		case float64:
			limit = int(v)
		case int:
			limit = v
		}
	}

	// Create FileTools and read
	ft := &skills.FileTools{}
	content, err := ft.ReadFile(path, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]any{
		"content": content,
	}, nil
}

func (t *readFileTool) ProcessRequest(_ tool.Context, req *model.LLMRequest) error {
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

// writeFileTool writes content to a file.
type writeFileTool struct{}

// NewWriteFileTool creates a new write file tool.
func NewWriteFileTool() tool.Tool {
	return &writeFileTool{}
}

func (t *writeFileTool) Name() string {
	return "write_file"
}

func (t *writeFileTool) Description() string {
	return "Write content to a file, creating parent directories if needed. Use this to create new files or completely replace existing file contents."
}

func (t *writeFileTool) IsLongRunning() bool {
	return false
}

func (t *writeFileTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: "OBJECT",
			Properties: map[string]*genai.Schema{
				"path": {
					Type:        "STRING",
					Description: "Path to the file to write",
				},
				"content": {
					Type:        "STRING",
					Description: "Content to write to the file",
				},
			},
			Required: []string{"path", "content"},
		},
	}
}

func (t *writeFileTool) Run(_ tool.Context, args any) (map[string]any, error) {
	m, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected args type, got: %T", args)
	}

	pathRaw, exists := m["path"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: path")
	}

	path, ok := pathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string, got: %T", pathRaw)
	}

	contentRaw, exists := m["content"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: content")
	}

	content, ok := contentRaw.(string)
	if !ok {
		return nil, fmt.Errorf("content must be a string, got: %T", contentRaw)
	}

	// Create FileTools and write
	ft := &skills.FileTools{}
	if err := ft.WriteFile(path, content); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]any{
		"status": fmt.Sprintf("Successfully wrote to %s", path),
	}, nil
}

func (t *writeFileTool) ProcessRequest(_ tool.Context, req *model.LLMRequest) error {
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

// editFileTool performs an exact string replacement in a file.
type editFileTool struct{}

// NewEditFileTool creates a new edit file tool.
func NewEditFileTool() tool.Tool {
	return &editFileTool{}
}

func (t *editFileTool) Name() string {
	return "edit_file"
}

func (t *editFileTool) Description() string {
	return "Edit a file by performing an exact string replacement. Use this to make precise edits to existing files. For creating new files or complete replacements, use write_file instead."
}

func (t *editFileTool) IsLongRunning() bool {
	return false
}

func (t *editFileTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: "OBJECT",
			Properties: map[string]*genai.Schema{
				"path": {
					Type:        "STRING",
					Description: "Path to the file to edit",
				},
				"old_string": {
					Type:        "STRING",
					Description: "The exact string to replace (must exist in the file)",
				},
				"new_string": {
					Type:        "STRING",
					Description: "The replacement string",
				},
				"replace_all": {
					Type:        "BOOLEAN",
					Description: "If true, replace all occurrences. If false (default), only the first occurrence is replaced",
				},
			},
			Required: []string{"path", "old_string", "new_string"},
		},
	}
}

func (t *editFileTool) Run(_ tool.Context, args any) (map[string]any, error) {
	m, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected args type, got: %T", args)
	}

	pathRaw, exists := m["path"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: path")
	}

	path, ok := pathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string, got: %T", pathRaw)
	}

	oldStringRaw, exists := m["old_string"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: old_string")
	}

	oldString, ok := oldStringRaw.(string)
	if !ok {
		return nil, fmt.Errorf("old_string must be a string, got: %T", oldStringRaw)
	}

	newStringRaw, exists := m["new_string"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: new_string")
	}

	newString, ok := newStringRaw.(string)
	if !ok {
		return nil, fmt.Errorf("new_string must be a string, got: %T", newStringRaw)
	}

	// Extract optional replace_all parameter (default to false)
	var replaceAll bool
	if replaceAllRaw, exists := m["replace_all"]; exists {
		replaceAll, ok = replaceAllRaw.(bool)
		if !ok {
			return nil, fmt.Errorf("replace_all must be a boolean, got: %T", replaceAllRaw)
		}
	}

	// Create FileTools and edit
	ft := &skills.FileTools{}
	if err := ft.EditFile(path, oldString, newString, replaceAll); err != nil {
		return nil, fmt.Errorf("failed to edit file: %w", err)
	}

	return map[string]any{
		"status": fmt.Sprintf("Successfully edited %s", path),
	}, nil
}

func (t *editFileTool) ProcessRequest(_ tool.Context, req *model.LLMRequest) error {
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
