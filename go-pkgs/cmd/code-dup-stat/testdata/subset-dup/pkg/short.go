package pkg

func DoQuick(data []int) int {
	sum := 0
	for _, v := range data {
		sum += v
	}
	return sum
}
