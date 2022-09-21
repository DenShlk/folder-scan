package util

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func AbsInt64(a int64) int64 {
	if a >= 0 {
		return a
	}
	return -a
}
