package fui

func typeFormatTable(v any) string {
	switch v.(type) {
	case *int:
		return "%d"
	case *float64:
		return "%.5g"
	case *rune:
		return "%c"
	default:
		return "%+v"
	}
}
