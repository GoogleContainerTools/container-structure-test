package strftime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombine(t *testing.T) {
	{
		s, _ := New(`%A foo`)
		if !assert.Equal(t, 1, len(s.compiled), "there are 1 element") {
			return
		}
	}
	{
		s, _ := New(`%A 100`)
		if !assert.Equal(t, 2, len(s.compiled), "there are two elements") {
			return
		}
	}
	{
		s, _ := New(`%A Mon`)
		if !assert.Equal(t, 2, len(s.compiled), "there are two elements") {
			return
		}
	}
}
