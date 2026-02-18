package numbers

// RoundToDecimalPlaces rounds a float32 value to the specified number of decimal places.
func RoundToDecimalPlaces(value float32, precision uint8) float32 {
	multiplier := float32(1)
	for i := uint8(0); i < precision; i++ {
		multiplier *= 10
	}

	// Add 0.5 for rounding and truncate
	if value >= 0 {
		return float32(int32(value*multiplier+0.5)) / multiplier
	}
	return float32(int32(value*multiplier-0.5)) / multiplier
}

// Scale multiplies a value by a scaling factor and rounds to the specified precision (default: 2).
// Useful for scaling recipe quantities when adjusting serving sizes.
// For example, Scale(2.5, 2.0) would return 5.0 (doubling the quantity).
func Scale(value, factor float32, precision ...uint8) float32 {
	result := value * factor

	p := uint8(2)
	if len(precision) > 0 {
		p = precision[0]
	}

	return RoundToDecimalPlaces(result, p)
}

// ScaleToYield scales a recipe quantity from an original yield to a desired yield.
// The optional precision parameter specifies the number of decimal places to round to (default: 2).
// For example, ScaleToYield(2.0, 4, 6) would return 3.0 (scaling 2 cups from 4 servings to 6 servings).
func ScaleToYield(originalValue float32, originalYield, desiredYield int, precision ...uint8) float32 {
	if originalYield <= 0 {
		return originalValue
	}

	factor := float32(desiredYield) / float32(originalYield)
	return Scale(originalValue, factor, precision...)
}
