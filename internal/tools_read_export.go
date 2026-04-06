package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerReadExportTools(s *server.MCPServer, node *Node) {
	s.AddTool(mcp.NewTool("get_screenshot",
		mcp.WithDescription("Export a screenshot of selected or specific nodes. Returns base64-encoded image data."),
		mcp.WithArray("nodeIds",
			mcp.Description("Optional node IDs to export, colon format. If empty, exports current selection."),
			mcp.WithStringItems(),
		),
		mcp.WithString("format",
			mcp.Description("Export format: PNG (default), SVG, JPG, or PDF"),
		),
		mcp.WithNumber("scale",
			mcp.Description("Export scale for raster formats (default 2)"),
		),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		raw, _ := req.GetArguments()["nodeIds"].([]interface{})
		nodeIDs := toStringSlice(raw)
		params := map[string]interface{}{}
		if f, ok := req.GetArguments()["format"].(string); ok && f != "" {
			params["format"] = f
		}
		if s, ok := req.GetArguments()["scale"].(float64); ok && s > 0 {
			params["scale"] = s
		}
		resp, err := node.Send(ctx, "get_screenshot", nodeIDs, params)
		return renderResponse(resp, err)
	})

	s.AddTool(mcp.NewTool("export_frames_to_pdf",
		mcp.WithDescription("Export multiple frames as a single multi-page PDF file. Each frame becomes one page in order. Ideal for pitch decks, proposals, and slide exports."),
		mcp.WithArray("nodeIds",
			mcp.Required(),
			mcp.Description("Ordered list of frame node IDs to export as PDF pages, colon format e.g. '4029:12345'"),
			mcp.WithStringItems(),
		),
		mcp.WithString("outputPath",
			mcp.Required(),
			mcp.Description("File path to write the PDF to, must end in .pdf (relative to working directory or absolute)"),
		),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		raw, _ := req.GetArguments()["nodeIds"].([]interface{})
		nodeIDs := toStringSlice(raw)
		outputPath, _ := req.GetArguments()["outputPath"].(string)
		if outputPath == "" {
			return mcp.NewToolResultError("outputPath is required"), nil
		}
		return executeExportFramesToPDF(ctx, node, nodeIDs, outputPath)
	})

	s.AddTool(mcp.NewTool("save_screenshots",
		mcp.WithDescription("Export screenshots for multiple nodes and save them to the local filesystem. Returns metadata only (no base64)."),
		mcp.WithArray("items",
			mcp.Required(),
			mcp.Description("List of {nodeId, outputPath, format?, scale?} objects"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"nodeId":     map[string]any{"type": "string", "description": "Node ID in colon format e.g. '4029:12345'"},
					"outputPath": map[string]any{"type": "string", "description": "File path to write the image to"},
					"format":     map[string]any{"type": "string", "description": "Export format: PNG, SVG, JPG, or PDF"},
					"scale":      map[string]any{"type": "number", "description": "Export scale for raster formats"},
				},
				"required": []string{"nodeId", "outputPath"},
			}),
		),
		mcp.WithString("format",
			mcp.Description("Default export format: PNG (default), SVG, JPG, or PDF"),
		),
		mcp.WithNumber("scale",
			mcp.Description("Default export scale for raster formats (default 2)"),
		),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return executeSaveScreenshots(ctx, node, req)
	})
}

func executeExportFramesToPDF(ctx context.Context, node *Node, nodeIDs []string, outputPath string) (*mcp.CallToolResult, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("getwd: %v", err)), nil
	}
	resolvedPath, err := resolveOutputPath(outputPath, workDir)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if strings.ToLower(filepath.Ext(resolvedPath)) != ".pdf" {
		return mcp.NewToolResultError("outputPath must have a .pdf extension"), nil
	}

	resp, err := node.Send(ctx, "export_frames_to_pdf", nodeIDs, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if resp.Error != "" {
		return mcp.NewToolResultError(resp.Error), nil
	}

	b64, err := extractPDFBase64(resp.Data)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	bytesWritten, err := writeBase64(b64, resolvedPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	out, _ := json.Marshal(map[string]interface{}{
		"outputPath":   resolvedPath,
		"bytesWritten": bytesWritten,
		"pageCount":    len(nodeIDs),
		"success":      true,
	})
	return mcp.NewToolResultText(string(out)), nil
}

func extractPDFBase64(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	var wrapper struct {
		Base64 string `json:"base64"`
	}
	if err := json.Unmarshal(b, &wrapper); err != nil {
		return "", err
	}
	if wrapper.Base64 == "" {
		return "", errors.New("no PDF data returned by plugin")
	}
	return wrapper.Base64, nil
}
