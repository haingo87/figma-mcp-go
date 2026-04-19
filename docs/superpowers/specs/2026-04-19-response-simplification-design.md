# Response Simplification Design

## Status

Approved — implementation pending.

## Overview

Port Framelink's response simplification layer from Figma-Context-MCP into figma-mcp-go. Raw Figma plugin API responses are transformed into a compact, LLM-friendly format before being returned to the LLM — reducing token usage and improving response accuracy.

## Background

figma-mcp-go returns raw Figma plugin API data directly to the LLM. This data is verbose and includes many fields irrelevant to code generation (e.g. internal IDs, deprecated fields, full paint metadata). Framelink (Figma-Context-MCP) demonstrated that transforming responses before sending to the LLM — stripping irrelevant fields, restructuring styles into CSS-like format — dramatically improves AI accuracy and reduces token consumption.

## Design

### Integration Point

The transform runs in `tools.go` in `renderResponse()`, after `node.Send()` returns and before `json.Marshal()`:

```
BridgeResponse.Data (raw Figma plugin data)
       ↓
  Simplify(data)       internal/transform/simplify.go
       ↓
  Simplified JSON      →  mcp.CallToolResult text
```

### New Package: internal/transform/

```
internal/transform/
  types.go              SimplifiedNode, SimplifiedDesign, GlobalVars, TraversalContext
  layout.go             buildSimplifiedLayout() — CSS flex-like layout schema
  style.go              parsePaint(), buildSimplifiedStrokes() — CSS colors/shadows
  text.go               buildFormattedText(), extractTextStyle() — plain text + typography
  effects.go            buildSimplifiedEffects() — box-shadow, blur
  component.go          simplifyComponentProperties(), etc.
  simplify.go           Simplify() — tree walker + all extractors wired up
  identity.go           hasValue(), isVisible(), isFrame() etc. (ported from Framelink utils)
  common.go             pixelRound(), generateVarId(), stableStringify() etc.
```

### API: Simplify(Data, Options)

```go
// Options controls which extractors run and traversal depth.
type Options struct {
    MaxDepth int
    // Which extractor combo to use:
    //   AllExtractors        — everything (Phase 1 default)
    //   LayoutAndText        — structure + text only
    //   VisualsOnly          — fills/strokes/effects only
    //   ContentOnly          — text only
    Extractors ExtractorCombo
}

// Simplify converts raw Figma plugin data into a SimplifiedDesign.
func Simplify(data interface{}, opts Options) (*SimplifiedDesign, error)
```

### Opt-in per Tool

Add `simplify` parameter to read tools. Existing behavior unchanged when `simplify` is omitted or false.

Tools that support `simplify`:
- `get_document`
- `get_design_context`
- `get_node`
- `get_nodes_info`
- `scan_text_nodes`
- `scan_nodes_by_types`
- `search_nodes`
- `get_selection`

### Plugin API Field Mapping

Framelink targets the Figma REST API. The plugin API uses slightly different field names:

| REST API | Plugin API | Notes |
|----------|-----------|-------|
| `absoluteBoundingBox` | `absoluteRenderBounds` | Node bounds |
| `itemSpacing` | `spacing` | Auto-layout gap |
| `effects` | same | Drop shadows, blurs |

All transformer code uses plugin API field names directly.

### Phase 1 Scope

Port the core extractors and transformers, skipping complex features:

**Included:**
- `layoutExtractor` + `layout.go` — layout mode (row/column/none), gap, padding, sizing, absolute positioning
- `visualsExtractor` + `style.go` + `effects.go` — fills as CSS (`rgba()`, linear-gradient, box-shadow), strokes, opacity, corner radius
- `textExtractor` + `text.go` — text content + basic typography (fontFamily, fontSize, fontWeight, lineHeight, letterSpacing). Plain text only, no inline style refs (`{ts1}**bold**{/ts1}`)
- `componentExtractor` + `component.go` — component IDs and properties (BOOLEAN/TEXT only)
- `node-walker` style tree traversal with depth limiting
- Deduplicated `globalVars` with `style_*` prefix for named styles

**Excluded for now:**
- Rich text inline style refs (`{ts1}**bold**{/ts1}`)
- SVG collapse (`IMAGE-SVG` type collapsing)
- OpenType flags, paragraph indent, list spacing
- Figma named style resolution (always auto-generate IDs)

### Output Format

A `SimplifiedDesign` containing:

```json
{
  "name": "Page 1",
  "nodes": [
    {
      "id": "4029:123",
      "name": "Header",
      "type": "FRAME",
      "layout": "layout_xxx",
      "fills": "fill_xxx",
      "text": "Welcome",
      "textStyle": "style_xxx",
      "opacity": 0.5,
      "borderRadius": "8px",
      "children": [...]
    }
  ],
  "components": { ... },
  "componentSets": { ... },
  "globalVars": {
    "layout_xxx": { "mode": "row", "gap": "8px", "padding": "16px" },
    "fill_xxx": "#3B82F6",
    "style_xxx": { "fontFamily": "Inter", "fontSize": 16, "fontWeight": 600 }
  }
}
```

### SimplifiedNode fields (Phase 1)

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Node ID |
| `name` | string | Node name |
| `type` | string | Figma node type (VECTOR → IMAGE-SVG) |
| `text` | string | Text content (plain) |
| `textStyle` | string | Reference to globalVars typography entry |
| `fills` | string | Reference to globalVars fill entry |
| `layout` | string | Reference to globalVars layout entry |
| `effects` | string | Reference to globalVars effects entry |
| `opacity` | float | Node opacity (only when ≠ 1) |
| `borderRadius` | string | CSS shorthand e.g. "8px" |
| `componentId` | string | Component ID (INSTANCE nodes only) |
| `componentProperties` | map | Simplified bool/string props (INSTANCE only) |
| `children` | []SimplifiedNode | Child nodes |

### Error Handling

- If `simplify` is true but simplification fails, fall back to returning raw JSON (never error out silently)
- Plugin not connected errors are returned as-is before any simplification is attempted

## Implementation Order

1. `internal/transform/types.go` — define all types
2. `internal/transform/common.go` + `identity.go` — utility functions
3. `internal/transform/layout.go` — layout transformer
4. `internal/transform/style.go` — paint/stroke transformer
5. `internal/transform/effects.go` — effects transformer
6. `internal/transform/text.go` — text transformer (plain text only)
7. `internal/transform/component.go` — component transformer
8. `internal/transform/simplify.go` — tree walker + extractors + public Simplify() API
9. Wire `simplify` parameter into `renderResponse()` and read tool handlers
10. Add tests for each transformer with plugin API data fixtures
