package main

func main() {
	sum := 0
	for _, p := range gen() {
		sum += *p
	}
	println(sum)
}

func gen() []*int {
	s := make([]*int, 0, 20)
	for i := range 100 {
		p := new(int)
		*p = i
		s = append(s, p)
	}
	return s
}
