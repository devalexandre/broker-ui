package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// DraculaTheme is a custom implementation of the theme for Fyne
type DraculaTheme struct{}

// Colors for the Dracula theme
var draculaColors = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:          color.RGBA{40, 42, 54, 255}, // Dark background
	theme.ColorNameButton:              color.RGBA{68, 71, 90, 255}, // Buttons
	theme.ColorNameDisabled:            color.RGBA{98, 114, 164, 255},
	theme.ColorNameDisabledButton:      color.RGBA{68, 71, 90, 255},
	theme.ColorNameError:               color.RGBA{255, 85, 85, 255}, // Error color
	theme.ColorNameFocus:               color.RGBA{189, 147, 249, 255},
	theme.ColorNameForeground:          color.RGBA{248, 248, 242, 255}, // Text
	theme.ColorNameForegroundOnError:   color.RGBA{255, 255, 255, 255},
	theme.ColorNameForegroundOnPrimary: color.RGBA{40, 42, 54, 255},
	theme.ColorNameForegroundOnSuccess: color.RGBA{40, 42, 54, 255},
	theme.ColorNameForegroundOnWarning: color.RGBA{40, 42, 54, 255},
	theme.ColorNameHeaderBackground:    color.RGBA{68, 71, 90, 255},
	theme.ColorNameHover:               color.RGBA{45, 47, 60, 255}, // Darker hover color for buttons
	theme.ColorNameHyperlink:           color.RGBA{139, 233, 253, 255},
	theme.ColorNameInputBackground:     color.RGBA{68, 71, 90, 255}, // Inputs
	theme.ColorNameInputBorder:         color.RGBA{98, 114, 164, 255},
	theme.ColorNameMenuBackground:      color.RGBA{50, 50, 62, 255},
	theme.ColorNameOverlayBackground:   color.RGBA{40, 42, 54, 255},
	theme.ColorNamePlaceHolder:         color.RGBA{98, 114, 164, 255},
	theme.ColorNamePressed:             color.RGBA{150, 50, 50, 255},   // Darker pressed color for error buttons
	theme.ColorNamePrimary:             color.RGBA{139, 233, 253, 255}, // Primary color
	theme.ColorNameScrollBar:           color.RGBA{68, 71, 90, 255},
	theme.ColorNameSelection:           color.RGBA{98, 114, 164, 255},
	theme.ColorNameSeparator:           color.RGBA{68, 71, 90, 255},
	theme.ColorNameShadow:              color.RGBA{0, 0, 0, 110},
	theme.ColorNameSuccess:             color.RGBA{80, 250, 123, 255},  // Success color
	theme.ColorNameWarning:             color.RGBA{255, 184, 108, 255}, // Warning color
}

// Implement the Color function for the theme
func (d DraculaTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := draculaColors[name]; ok {
		return c
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Implement the Font function for the theme
func (d DraculaTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Implement the Size function for the theme
func (d DraculaTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Implement the Icon function for the theme
func (d DraculaTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
