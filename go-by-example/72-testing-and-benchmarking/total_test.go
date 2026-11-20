package main

import (
	"testing"
)

func total(prices []int) int {
	sum := 0
	for _, price := range prices {
		sum += price
	}
	return sum
}

func TestTotal(t *testing.T) {
	cases := []struct {
		prices []int
		want   int
	}{{nil, 0}, {[]int{2, 5}, 7}, {[]int{4, -1}, 3}}
	for _, tc := range cases {
		if got := total(tc.prices); got != tc.want {
			t.Errorf("total(%v) = %d, want %d", tc.prices, got, tc.want)
		}
	}
}

func BenchmarkTotal(b *testing.B) {
	prices := []int{2, 5, 8}
	for b.Loop() {
		total(prices)
	}
}
