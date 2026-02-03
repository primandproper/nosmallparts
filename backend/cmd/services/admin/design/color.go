package design

type Color struct {
	Name  string
	Value string
}

type Palette struct {
	Primary    Color
	Secondary  Color
	Warning    Color
	Accent     Color
	Background Color
	Text       Color
}

var StandardPalette = Palette{
	Primary:    Color{Name: "Primary Blue", Value: "blue-500"},
	Secondary:  Color{Name: "Secondary Pink", Value: "pink-400"},
	Accent:     Color{Name: "Accent Yellow", Value: "yellow-300"},
	Warning:    Color{Name: "Warning Red", Value: "red-600"},
	Background: Color{Name: "Background Gray", Value: "gray-100"},
	Text:       Color{Name: "Text Dark", Value: "gray-700"},
}

func Background(c Color) string {
	return "bg-" + c.Value
}

func TextColor(c Color) string {
	return "text-" + c.Value
}

func BorderColor(c Color) string {
	return "border-" + c.Value
}
