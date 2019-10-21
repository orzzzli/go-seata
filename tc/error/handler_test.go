package error

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	a := New("a", 1, nil)
	b := New("b", 2, errors.New("some error"), "aaaa", 1)
	var temp []string
	temp = append(temp, "aaa")
	temp = append(temp, "bbb")
	temp = append(temp, "ccc")
	c := New("c", 3, errors.New("some error"), "a", 1, temp)
	println(a.ToStr())
	println(b.ToStr())
	println(c.ToStr())
	fmt.Println(c.ToReplyInfo())
}
