package prompts

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func addReadDesignStrategy(s *server.MCPServer) {
	s.AddPrompt(mcp.NewPrompt("read_design_strategy",
		mcp.WithPromptDescription("Best practices for reading Figma designs with figma-mcp-go"),
	), func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return mcp.NewGetPromptResult(
			"Best practices for reading Figma designs",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(`To effectively read a Figma design with figma-mcp-go:

1. Start with get_metadata — understand file name, pages, and current page
2. Use get_pages to list all pages without loading their full trees
3. Use get_design_context (depth=2, detail=compact) for a token-efficient summary of the current selection or page
   - detail=minimal: id/name/type/bounds only (~5% tokens)
   - detail=compact: + fills/strokes/opacity (~30% tokens)
   - detail=full: everything, default (100% tokens)
4. Use search_nodes to find nodes by name or type without dumping the entire tree
5. Drill into specific nodes with get_node or get_nodes_info (prefer batch over single calls)
6. For text-heavy components, use scan_text_nodes to collect all copy at once
7. Use scan_nodes_by_types to find all FRAME/COMPONENT/INSTANCE nodes in a subtree
8. Call get_styles and get_variable_defs once per session to understand the design system
9. Call get_fonts to understand typography usage across the page at a glance
10. Use get_viewport to see what the user is currently looking at in the canvas
11. Use get_reactions to inspect prototype interactions on a node
12. Call get_screenshot last and only when visual confirmation is needed — it is expensive
13. Node IDs use colon format: 4029:12345 — never use hyphens
14. get_local_components now includes componentSets and variantProperties for variant-aware inspection`),
				),
			},
		), nil
	})
}
