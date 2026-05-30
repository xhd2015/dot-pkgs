package pkg

func DoFull(data []int) int {
	sum := 0
	for _, v := range data {
		sum += v
	}
	if sum > 100 {
		sum = sum / 2
	}
	if sum < 0 {
		sum = 0
	}
	return sum
}
