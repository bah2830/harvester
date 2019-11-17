package main

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
)

type baseTheme struct {
	background color.Color

	button, text, icon, hyperlink, placeholder, primary, hover, scrollBar, shadow color.Color
	regular, bold, italic, bolditalic, monospace                                  fyne.Resource
	disabledButton, disabledIcon, disabledText                                    color.Color
}

func defaultTheme() fyne.Theme {
	dt := theme.DarkTheme()

	return &baseTheme{
		background:     color.RGBA{0x1a, 0x1d, 0x21, 0xff},
		button:         color.RGBA{0x29, 0x2c, 0x2f, 0xff},
		disabledButton: color.RGBA{0x31, 0x31, 0x31, 0xff},
		text:           color.RGBA{0xbc, 0xbe, 0xbf, 0xff},
		disabledText:   color.RGBA{0x60, 0x60, 0x60, 0xff},
		icon:           color.RGBA{0xbc, 0xbe, 0xbf, 0xff},
		disabledIcon:   color.RGBA{0x60, 0x60, 0x60, 0xff},
		hyperlink:      color.RGBA{0xbc, 0xbe, 0xbf, 0xff},
		placeholder:    color.RGBA{0xb2, 0xb2, 0xb2, 0xff},
		primary:        color.RGBA{0x1a, 0x23, 0x7e, 0xff},
		hover:          color.RGBA{0x39, 0x3c, 0x3f, 0xff},
		scrollBar:      color.RGBA{0x0, 0x0, 0x0, 0x99},
		shadow:         color.RGBA{0x0, 0x0, 0x0, 0x66},
		regular:        dt.TextFont(),
		bold:           dt.TextBoldFont(),
		italic:         dt.TextItalicFont(),
		bolditalic:     dt.TextBoldItalicFont(),
		monospace:      dt.TextMonospaceFont(),
	}
}

func (t *baseTheme) BackgroundColor() color.Color {
	return t.background
}

// ButtonColor returns the theme's standard button colour
func (t *baseTheme) ButtonColor() color.Color {
	return t.button
}

// DisabledButtonColor returns the theme's disabled button colour
func (t *baseTheme) DisabledButtonColor() color.Color {
	return t.disabledButton
}

// HyperlinkColor returns the theme's standard hyperlink colour
func (t *baseTheme) HyperlinkColor() color.Color {
	return t.hyperlink
}

// TextColor returns the theme's standard text colour
func (t *baseTheme) TextColor() color.Color {
	return t.text
}

// DisabledIconColor returns the color for a disabledIcon UI element
func (t *baseTheme) DisabledTextColor() color.Color {
	return t.disabledText
}

// IconColor returns the theme's standard text colour
func (t *baseTheme) IconColor() color.Color {
	return t.icon
}

// DisabledIconColor returns the color for a disabledIcon UI element
func (t *baseTheme) DisabledIconColor() color.Color {
	return t.disabledIcon
}

// PlaceHolderColor returns the theme's placeholder text colour
func (t *baseTheme) PlaceHolderColor() color.Color {
	return t.placeholder
}

// PrimaryColor returns the colour used to highlight primary features
func (t *baseTheme) PrimaryColor() color.Color {
	return t.primary
}

// HoverColor returns the colour used to highlight interactive elements currently under a cursor
func (t *baseTheme) HoverColor() color.Color {
	return t.hover
}

// FocusColor returns the colour used to highlight a focused widget
func (t *baseTheme) FocusColor() color.Color {
	return t.primary
}

// ScrollBarColor returns the color (and translucency) for a scrollBar
func (t *baseTheme) ScrollBarColor() color.Color {
	return t.scrollBar
}

// ShadowColor returns the color (and translucency) for shadows used for indicating elevation
func (t *baseTheme) ShadowColor() color.Color {
	return t.shadow
}

// TextSize returns the standard text size
func (t *baseTheme) TextSize() int {
	return 12
}

// TextFont returns the font resource for the regular font style
func (t *baseTheme) TextFont() fyne.Resource {
	return t.regular
}

// TextBoldFont retutns the font resource for the bold font style
func (t *baseTheme) TextBoldFont() fyne.Resource {
	return t.bold
}

// TextItalicFont returns the font resource for the italic font style
func (t *baseTheme) TextItalicFont() fyne.Resource {
	return t.italic
}

// TextBoldItalicFont returns the font resource for the bold and italic font style
func (t *baseTheme) TextBoldItalicFont() fyne.Resource {
	return t.bolditalic
}

// TextMonospaceFont retutns the font resource for the monospace font face
func (t *baseTheme) TextMonospaceFont() fyne.Resource {
	return t.monospace
}

// Padding is the standard gap between elements and the border around interface
// elements
func (t *baseTheme) Padding() int {
	return 6
}

// IconInlineSize is the standard size of icons which appear within buttons, labels etc.
func (t *baseTheme) IconInlineSize() int {
	return 20
}

// ScrollBarSize is the width (or height) of the bars on a ScrollContainer
func (t *baseTheme) ScrollBarSize() int {
	return 16
}

// ScrollBarSmallSize is the width (or height) of the minimized bars on a ScrollContainer
func (t *baseTheme) ScrollBarSmallSize() int {
	return 3
}
