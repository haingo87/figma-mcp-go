package transform

// buildSimplifiedLayout extracts layout mode, gap, padding, sizing from a raw node map.
// Plugin API field names: spacing (not itemSpacing), absoluteRenderBounds (not absoluteBoundingBox).
func buildSimplifiedLayout(node map[string]any, ctx *TraversalContext) (string, *SimplifiedLayout) {
	layoutMode := getString(node, "layoutMode")
	if layoutMode == "" && !isFrame(node) {
		return "", nil
	}

	layout := &SimplifiedLayout{}

	// Mode
	if layoutMode == "HORIZONTAL" {
		layout.Mode = "row"
	} else if layoutMode == "VERTICAL" {
		layout.Mode = "column"
	} else {
		layout.Mode = "none"
	}

	// Spacing (gap) — plugin uses "spacing", not "itemSpacing"
	if spacing := getFloat64(node, "spacing"); spacing > 0 {
		layout.Gap = cssLen(spacing, "px")
	}

	// Item spacing (horizontal gap in row mode)
	if itemSpacing := getFloat64(node, "itemSpacing"); itemSpacing > 0 {
		layout.ItemSpacing = cssLen(itemSpacing, "px")
	}

	// Padding
	paddingTop := getFloat64(node, "paddingTop")
	paddingRight := getFloat64(node, "paddingRight")
	paddingBottom := getFloat64(node, "paddingBottom")
	paddingLeft := getFloat64(node, "paddingLeft")
	if paddingTop > 0 || paddingRight > 0 || paddingBottom > 0 || paddingLeft > 0 {
		layout.PaddingTop = cssLen(paddingTop, "px")
		layout.PaddingRight = cssLen(paddingRight, "px")
		layout.PaddingBottom = cssLen(paddingBottom, "px")
		layout.PaddingLeft = cssLen(paddingLeft, "px")
	}

	// Sizing — primaryAxisSizingMode / counterAxisSizingMode
	primarySizing := getString(node, "primaryAxisSizingMode")
	counterSizing := getString(node, "counterAxisSizingMode")
	if primarySizing == "FIXED" {
		layout.Width = "fixed"
	} else if primarySizing == "AUTO" {
		layout.Width = "hug"
	}
	if counterSizing == "FIXED" {
		layout.Height = "fixed"
	} else if counterSizing == "AUTO" {
		layout.Height = "hug"
	}

	// Absolute positioning — no layout mode
	if layoutMode == "NONE" {
		layout.Mode = "none"
	}

	if layout.Mode == "none" && layout.Gap == "" && layout.PaddingTop == "" {
		return "", nil
	}

	ctx.IDCounters.Layout++
	id := generateVarId("layout", ctx.IDCounters.Layout)
	return id, layout
}