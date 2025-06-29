package light

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// LightTheme é uma implementação personalizada do tema claro para Fyne
type LightTheme struct{}

// Colors do tema Light
var lightColors = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.RGBA{248, 249, 250, 255}, // Fundo claro
	theme.ColorNameButton:          color.RGBA{233, 236, 239, 255}, // Botões
	theme.ColorNameDisabled:        color.RGBA{173, 181, 189, 255},
	theme.ColorNameDisabledButton:  color.RGBA{233, 236, 239, 255},
	theme.ColorNameError:           color.RGBA{220, 53, 69, 255},   // Vermelho para erros
	theme.ColorNameForeground:      color.RGBA{33, 37, 41, 255},    // Texto escuro
	theme.ColorNameHover:           color.RGBA{222, 226, 230, 255}, // Cor de hover
	theme.ColorNameInputBackground: color.RGBA{255, 255, 255, 255}, // Inputs brancos
	theme.ColorNamePlaceHolder:     color.RGBA{108, 117, 125, 255}, // Placeholder cinza
	theme.ColorNamePrimary:         color.RGBA{13, 110, 253, 255},  // Azul primário
	theme.ColorNameScrollBar:       color.RGBA{206, 212, 218, 255}, // ScrollBar
	theme.ColorNameShadow:          color.RGBA{0, 0, 0, 25},        // Sombra suave
}

// Implementa a função de cor para o tema
func (l LightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := lightColors[name]; ok {
		return c
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Implementa a função de fonte para o tema
func (l LightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Implementa a função de tamanho para o tema
func (l LightTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Implementa a função de ícone para o tema
func (l LightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
