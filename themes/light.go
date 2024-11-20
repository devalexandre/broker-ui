package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// LightTheme is a custom implementation of a light theme for Fyne
type LightTheme struct{}

// Colors for the light theme
var lightColors = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.RGBA{255, 255, 255, 255}, // White background
	theme.ColorNameButton:          color.RGBA{240, 240, 240, 255}, // Light gray buttons
	theme.ColorNameDisabled:        color.RGBA{200, 200, 200, 255}, // Disabled elements
	theme.ColorNameDisabledButton:  color.RGBA{220, 220, 220, 255}, // Disabled buttons
	theme.ColorNameError:           color.RGBA{255, 0, 0, 255},     // Red error messages
	theme.ColorNameForeground:      color.RGBA{0, 0, 0, 255},       // Black text
	theme.ColorNameHover:           color.RGBA{230, 230, 230, 255}, // Hover effect
	theme.ColorNameInputBackground: color.RGBA{255, 255, 255, 255}, // Input fields
	theme.ColorNamePlaceHolder:     color.RGBA{160, 160, 160, 255}, // Placeholder text
	theme.ColorNamePrimary:         color.RGBA{0, 122, 204, 255},   // Primary color (blue)
	theme.ColorNameScrollBar:       color.RGBA{220, 220, 220, 255}, // Scroll bars
	theme.ColorNameShadow:          color.RGBA{0, 0, 0, 50},        // Shadows
}

// Implement the Color function for the theme
func (l LightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := lightColors[name]; ok {
		return c
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Implement the Font function for the theme
func (l LightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Implement the Size function for the theme
func (l LightTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Implement the Icon function for the theme
func (l LightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
