package material

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// MaterialTheme is a modern, Material-inspired theme.
// If Variant is set to theme.VariantLight or theme.VariantDark it will override the system variant.
type MaterialTheme struct {
	Variant fyne.ThemeVariant
}

var materialLight = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.NRGBA{R: 0xFA, G: 0xFA, B: 0xFA, A: 0xFF},
	theme.ColorNameButton:          color.NRGBA{R: 0xF1, G: 0xF3, B: 0xF4, A: 0xFF},
	theme.ColorNameDisabled:        color.NRGBA{R: 0xB0, G: 0xB5, B: 0xBA, A: 0xFF},
	theme.ColorNameDisabledButton:  color.NRGBA{R: 0xE6, G: 0xE8, B: 0xEA, A: 0xFF},
	theme.ColorNameError:           color.NRGBA{R: 0xD9, G: 0x2D, B: 0x20, A: 0xFF},
	theme.ColorNameForeground:      color.NRGBA{R: 0x1F, G: 0x23, B: 0x28, A: 0xFF},
	theme.ColorNameHover:           color.NRGBA{R: 0xE8, G: 0xEA, B: 0xED, A: 0xFF},
	theme.ColorNameInputBackground: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
	theme.ColorNamePlaceHolder:     color.NRGBA{R: 0x6B, G: 0x72, B: 0x7A, A: 0xFF},
	theme.ColorNamePrimary:         color.NRGBA{R: 0x1A, G: 0x73, B: 0xE8, A: 0xFF}, // Google Blue
	theme.ColorNameScrollBar:       color.NRGBA{R: 0xD0, G: 0xD5, B: 0xDA, A: 0xFF},
	theme.ColorNameShadow:          color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x22},
}

var materialDark = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.NRGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xFF},
	theme.ColorNameButton:          color.NRGBA{R: 0x1C, G: 0x1C, B: 0x1E, A: 0xFF},
	theme.ColorNameDisabled:        color.NRGBA{R: 0x67, G: 0x6F, B: 0x78, A: 0xFF},
	theme.ColorNameDisabledButton:  color.NRGBA{R: 0x23, G: 0x23, B: 0x26, A: 0xFF},
	theme.ColorNameError:           color.NRGBA{R: 0xF2, G: 0xB8, B: 0xB5, A: 0xFF},
	theme.ColorNameForeground:      color.NRGBA{R: 0xE8, G: 0xEA, B: 0xED, A: 0xFF},
	theme.ColorNameHover:           color.NRGBA{R: 0x24, G: 0x24, B: 0x28, A: 0xFF},
	theme.ColorNameInputBackground: color.NRGBA{R: 0x1C, G: 0x1C, B: 0x1E, A: 0xFF},
	theme.ColorNamePlaceHolder:     color.NRGBA{R: 0x9A, G: 0xA0, B: 0xA6, A: 0xFF},
	theme.ColorNamePrimary:         color.NRGBA{R: 0x8A, G: 0xB4, B: 0xF8, A: 0xFF}, // Light blue
	theme.ColorNameScrollBar:       color.NRGBA{R: 0x2F, G: 0x33, B: 0x38, A: 0xFF},
	theme.ColorNameShadow:          color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x99},
}

func (m MaterialTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if m.Variant == theme.VariantLight || m.Variant == theme.VariantDark {
		variant = m.Variant
	}

	if variant == theme.VariantDark {
		if c, ok := materialDark[name]; ok {
			return c
		}
	} else {
		if c, ok := materialLight[name]; ok {
			return c
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m MaterialTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m MaterialTheme) Size(name fyne.ThemeSizeName) float32 {
	// Slightly more spacious defaults.
	switch name {
	case theme.SizeNamePadding:
		return theme.DefaultTheme().Size(name) + 2
	case theme.SizeNameInlineIcon:
		return theme.DefaultTheme().Size(name) + 2
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func (m MaterialTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
