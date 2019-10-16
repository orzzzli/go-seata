package main

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkTest2(b *testing.B) {
	b.ResetTimer()
	for i:= 0 ;i<b.N;i++{
		rand.Seed(time.Now().Unix())
		run2(rand.Intn(50)+1)
	}
}
