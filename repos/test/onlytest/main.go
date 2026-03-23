package main

type S struct {
	data [30]int
}

func newS() *S {
	return &S{
		data: [30]int{},
	}
}

func main() {
	print(newS().data[0])
}
