package dracula

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// DraculaTheme is a custom implementation of Dracula theme for Fyne
type DraculaTheme struct{}

// Dracula theme colors
var draculaColors = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.RGBA{40, 42, 54, 255}, // Dark background
	theme.ColorNameButton:          color.RGBA{68, 71, 90, 255}, // Buttons
	theme.ColorNameDisabled:        color.RGBA{98, 114, 164, 255},
	theme.ColorNameDisabledButton:  color.RGBA{68, 71, 90, 255},
	theme.ColorNameError:           color.RGBA{255, 85, 85, 255},
	theme.ColorNameForeground:      color.RGBA{248, 248, 242, 255}, // Text
	theme.ColorNameHover:           color.RGBA{50, 50, 62, 255},    // Hover color
	theme.ColorNameInputBackground: color.RGBA{68, 71, 90, 255},    // Inputs
	theme.ColorNamePlaceHolder:     color.RGBA{98, 114, 164, 255},
	theme.ColorNamePrimary:         color.RGBA{139, 233, 253, 255}, // Primary color
	theme.ColorNameScrollBar:       color.RGBA{68, 71, 90, 255},
	theme.ColorNameShadow:          color.RGBA{0, 0, 0, 110},
}

// Implements color function for the theme
func (d DraculaTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := draculaColors[name]; ok {
		return c
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Implements font function for the theme
func (d DraculaTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Implements size function for the theme
func (d DraculaTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Implements icon function for the theme
func (d DraculaTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
