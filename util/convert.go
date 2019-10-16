package util

import (
	"math"
	"strconv"
)

func Int64to32(number int64) (int, error) {
	tempStr := strconv.FormatInt(number, 10)
	return strconv.Atoi(tempStr)
}

func Int32to64(number int) (int64, error) {
	tempStr := strconv.Itoa(number)
	return strconv.ParseInt(tempStr, 10, 64)
}

func Int32to64bitStr(number int) string {
	negative := false
	if number < 0 {
		negative = true
		number = -number
	}
	var outputByte []byte
	baseStr := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	baseByte := []byte(baseStr)
	for {
		left := number % 62
		number = int(math.Floor(float64(number) / 62))
		outputByte = append(outputByte, baseByte[left])
		if number == 0 {
			break
		}
	}
	outputStr := string(outputByte)
	if negative {
		outputStr = "-" + outputStr
	}
	return outputStr
}