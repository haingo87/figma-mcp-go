package internal

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerWriteVariableTools(s *server.MCPServer, node *Node) {
	s.AddTool(mcp.NewTool("create_variable_collection",
		mcp.WithDescription("Create a new local variable collection."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Collection name"),
		),
		mcp.WithString("initialModeName", mcp.Description("Name for the initial mode (default 'Mode 1')")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := req.GetArguments()
		resp, err := node.Send(ctx, "create_variable_collection", nil, params)
		return renderResponse(resp, err)
	})

	s.AddTool(mcp.NewTool("add_variable_mode",
		mcp.WithDescription("Add a new mode to an existing variable collection (e.g. Light/Dark, Desktop/Mobile)."),
		mcp.WithString("collectionId",
			mcp.Required(),
			mcp.Description("Variable collection ID"),
		),
		mcp.WithString("modeName",
			mcp.Required(),
			mcp.Description("Name for the new mode"),
		),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := req.GetArguments()
		resp, err := node.Send(ctx, "add_variable_mode", nil, params)
		return renderResponse(resp, err)
	})

	s.AddTool(mcp.NewTool("create_variable",
		mcp.WithDescription("Create a new variable (design token) inside an existing collection. Returns the new variable's ID. Use get_variable_defs to find collection IDs, set_variable_value to set values per mode, and bind_variable_to_node to apply the variable to a node property."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Variable name — use slash notation to group e.g. 'Color/Primary', 'Spacing/MD'"),
		),
		mcp.WithString("collectionId",
			mcp.Required(),
			mcp.Description("ID of the variable collection to add this variable to (from get_variable_defs)"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Variable type: COLOR (hex color), FLOAT (numeric dimension/spacing), STRING (text), or BOOLEAN (true/false toggle)"),
		),
		mcp.WithString("value", mcp.Description("Initial value for the first mode. COLOR: hex e.g. #FF5733. FLOAT: number e.g. 16. STRING: text. BOOLEAN: true or false.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := req.GetArguments()
		resp, err := node.Send(ctx, "create_variable", nil, params)
		return renderResponse(resp, err)
	})

	s.AddTool(mcp.NewTool("set_variable_value",
		mcp.WithDescription("Set a variable's value for a specific mode."),
		mcp.WithString("variableId",
			mcp.Required(),
			mcp.Description("Variable ID"),
		),
		mcp.WithString("modeId",
			mcp.Required(),
			mcp.Description("Mode ID within the collection"),
		),
		mcp.WithString("value",
			mcp.Required(),
			mcp.Description("Value to set. COLOR: hex e.g. #FF5733. FLOAT: number e.g. 16. STRING: text. BOOLEAN: true or false."),
		),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := req.GetArguments()
		resp, err := node.Send(ctx, "set_variable_value", nil, params)
		return renderResponse(resp, err)
	})

	s.AddTool(mcp.NewTool("delete_variable",
		mcp.WithDescription("Delete a single variable (provide variableId) or an entire collection and all its variables (provide collectionId). Provide exactly one of the two — not both."),
		mcp.WithString("variableId", mcp.Description("Variable ID to delete")),
		mcp.WithString("collectionId", mcp.Description("Collection ID to delete (removes all variables in the collection)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := req.GetArguments()
		resp, err := node.Send(ctx, "delete_variable", nil, params)
		return renderResponse(resp, err)
	})
}
