package light

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// LightTheme is a custom implementation of light theme for Fyne
type LightTheme struct{}

// Light theme colors
var lightColors = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.RGBA{248, 249, 250, 255}, // Light background
	theme.ColorNameButton:          color.RGBA{233, 236, 239, 255}, // Buttons
	theme.ColorNameDisabled:        color.RGBA{173, 181, 189, 255},
	theme.ColorNameDisabledButton:  color.RGBA{233, 236, 239, 255},
	theme.ColorNameError:           color.RGBA{220, 53, 69, 255},   // Red for errors
	theme.ColorNameForeground:      color.RGBA{33, 37, 41, 255},    // Dark text
	theme.ColorNameHover:           color.RGBA{222, 226, 230, 255}, // Hover color
	theme.ColorNameInputBackground: color.RGBA{255, 255, 255, 255}, // White inputs
	theme.ColorNamePlaceHolder:     color.RGBA{108, 117, 125, 255}, // Gray placeholder
	theme.ColorNamePrimary:         color.RGBA{13, 110, 253, 255},  // Primary blue
	theme.ColorNameScrollBar:       color.RGBA{206, 212, 218, 255}, // ScrollBar
	theme.ColorNameShadow:          color.RGBA{0, 0, 0, 25},        // Soft shadow
}

// Implements color function for the theme
func (l LightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := lightColors[name]; ok {
		return c
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Implements font function for the theme
func (l LightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Implements size function for the theme
func (l LightTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Implements icon function for the theme
func (l LightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
