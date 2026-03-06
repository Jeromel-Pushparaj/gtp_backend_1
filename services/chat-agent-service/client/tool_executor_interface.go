package client

// ToolExecutorInterface defines the interface for tool execution
type ToolExecutorInterface interface {
	ExecuteTool(toolName string, arguments map[string]interface{}) (string, error)
}