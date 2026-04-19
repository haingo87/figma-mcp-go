package transform

import "fmt"

// buildSimplifiedEffects extracts effects (shadows, blurs) into CSS box-shadow strings.
func buildSimplifiedEffects(node map[string]any, ctx *TraversalContext) (string, any) {
	effects := getSlice(node, "effects")
	if len(effects) == 0 {
		return "", nil
	}

	var shadows []string
	for _, e := range effects {
		if effect, ok := e.(map[string]any); ok {
			if css := parseEffect(effect); css != "" {
				shadows = append(shadows, css)
			}
		}
	}
	if len(shadows) == 0 {
		return "", nil
	}

	ctx.IDCounters.Effect++
	id := generateVarId("effect", ctx.IDCounters.Effect)

	var val any
	if len(shadows) == 1 {
		val = shadows[0]
	} else {
		val = shadows
	}
	ctx.Vars.Effects[id] = val
	return id, val
}

// parseEffect converts a single Figma effect to a CSS box-shadow / blur string.
func parseEffect(effect map[string]any) string {
	effectType, _ := effect["type"].(string)

	switch effectType {
	case "DROP_SHADOW", "INNER_SHADOW":
		color, _ := effect["color"].(map[string]any)
		opacity := 1.0
		if op, ok := effect["opacity"].(float64); ok {
			opacity = op
		}

		var r, g, b float64
		if color != nil {
			r = getFloat64(color, "r") * 255
			g = getFloat64(color, "g") * 255
			b = getFloat64(color, "b") * 255
		}

		offsetX := getFloat64(effect, "offsetX")
		offsetY := getFloat64(effect, "offsetY")
		blur := getFloat64(effect, "blurRadius")
		spread := getFloat64(effect, "spread")

		shadowType := "0"
		if effectType == "INNER_SHADOW" {
			shadowType = "inset "
		}

		return fmt.Sprintf("%s%.2fpx %.2fpx %.2fpx %.2fpx rgba(%d,%d,%d,%.2f)",
			shadowType, offsetX, offsetY, blur, spread,
			int(pixelRound(r)), int(pixelRound(g)), int(pixelRound(b)), opacity)

	case "LAYER_BLUR", "BACKGROUND_BLUR":
		blur := getFloat64(effect, "blurRadius")
		if blur > 0 {
			return fmt.Sprintf("blur(%.2fpx)", blur)
		}
	}
	return ""
}