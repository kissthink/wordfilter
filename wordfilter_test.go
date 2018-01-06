package wordfilter

import (
	"fmt"
	"testing"
)

func TestCheck(t *testing.T) {
	root := NewNode("")
	fb := NewSensitiveWordFilterBase(root)
	if fb.Load("word.txt") == false {
		fmt.Println("Load err")
		return
	}
	// check
	a := fb.Check("毛;\\泽;东你好吗")
	if a {
		t.Error("no equal a:", a)
	}
}

func BenchmarkCheck(b *testing.B) {
	root := NewNode("")
	fb := NewSensitiveWordFilterBase(root)
	if fb.Load("word.txt") == false {
		fmt.Println("Load err")
		return
	}

	for i := 0; i < b.N; i++ {
		// check
		a := fb.Check("毛;\\泽;东你好吗")
		if a {
			b.Error("no equal a:", a)
		}
	}
}

func TestFilter(t *testing.T) {
	root := NewNode("")
	fb := NewSensitiveWordFilterBase(root)
	if fb.Load("word.txt") == false {
		fmt.Println("Load err")
		return
	}

	// filter
	ss := fb.Filter("毛泽;东你吗")
	if ss != "**;你吗" {
		t.Error("no equal ss:", ss)
	}
}

func BenchmarkFilter(b *testing.B) {
	root := NewNode("")
	fb := NewSensitiveWordFilterBase(root)
	if fb.Load("word.txt") == false {
		fmt.Println("Load err")
		return
	}
	for i := 0; i < b.N; i++ {
		// filter
		ss := fb.Filter("毛泽;东你吗")
		if ss != "**;你吗" {
			b.Error("no equal ss:", ss)
		}
	}
}
