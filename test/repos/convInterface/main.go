package main

import (
	"fmt"
	"strconv"
)

func main() {
	assign()
	conv()
	call()
	send()
	ret()
}

func assign() {
	var i interface{}
	ch := make(chan int)
	go func() {
		ch <- 1
	}()
	i = <-ch
	println(i)
}

func assignToEface() {
	var eface Eface
	var iface Iface
	var i2 interface{}
	eface = true
	a := 1
	eface = a
	str := "abcdef"
	eface = str
	arr := [2]int{1, 2}
	eface = arr
	slice := []int{1, 2, 3}
	eface = slice
	s := S{
		a: 2,
	}
	eface = s
	ptr := &a
	eface = ptr
	f := func() {
		println(a)
		a = 2
		str = "fdas"
		println(str)
	}
	eface = f
	i2 = s
	eface = i2
	eface = iface
	m := map[int]string{
		1: "1",
		2: "2",
	}
	eface = m
	ch := make(chan bool, 2)
	eface = ch
	println(eface)
}

func assignToIface() {
	var iface Iface
	s := S{
		a: 2,
	}
	iface = s
	println(iface)
}

type Iface interface {
	String() string
}

type Eface interface{}

type S struct {
	a int
}

func (s S) String() string {
	return strconv.Itoa(s.a)
}

func conv() {
	var i interface{} = interface{}(S{
		a: 2,
	})
	println(i)
}

func call() {
	a := 1
	b := 2
	fmt.Println(a, b)
	s := []int{2, 3, 4}
	fmt.Println(s)
	s = append(s, []int{2, 3, 4}...)
}

func send() {
	ch := make(chan interface{})
	a := 2
	go func() {
		i := <-ch
		println(i)
	}()
	ch <- a
}

func ret() interface{} {
	a := 1
	return a
}
