package main

type B struct {
	f X
}

type X struct {
	g *int
}

func main() {
	a := 2
	var b B
	b.f.g = &a   // b.f.g 16,17,18 b.f 15,17
	x := b.f     // b.f 17
	print(x.g)   // x.g 17, 18
	print(b.f.g) // b.f.g 18
	print(x.g)
	d := 3
	var f2 X = X{
		g: &d,
	}
	b.f = f2 // should not be aliases for 15 b.f
	print(b.f)
}
