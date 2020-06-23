package schedule

// OptionFunc can be used customize a new Schedule.
type OptionFunc func(*Schedule) error

// Skip disables the skip or not.
func Skip(s bool) OptionFunc {
	return func(sc *Schedule) error {
		sc.Input.Skip = true
		return nil
	}
}

// Verbosity Print more if > 0
func Verbosity(v int) OptionFunc {
	return func(sc *Schedule) error {
		sc.Input.Verbosity = v
		return nil
	}
}

// Desc disables the phase description
func Desc(d string) OptionFunc {
	return func(sc *Schedule) error {
		sc.Input.Desc = d
		return nil
	}
}
