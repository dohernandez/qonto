package model

// Cents represents cents.
type Cents int64

// ToCents convert float into cents.
func ToCents(f float64) Cents {
	return Cents(f * 100)
}
