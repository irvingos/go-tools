package funcx

func IfElse[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func IfElseFunc[T any](condition bool, t func() T, f func() T) T {
	if condition {
		return t()
	}
	return f()
}
