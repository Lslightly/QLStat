package main

import "strconv"

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

func main() {
	assignToEface()
	assignToIface()
}

func assignToEface() {
	var eface Eface
	var i2 interface{}
	var iface Iface
	eface = iface
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
