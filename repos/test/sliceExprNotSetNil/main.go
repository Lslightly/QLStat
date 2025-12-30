package main

func main() {
	good()
	bad()
	goodLog()
	badLog()
}

type S struct{}
type sliceS []*S

type ptr *int
type slice []ptr

type Log struct {
	buffer slice
}

func good() {
	var s slice
	for i := 0; i < 8; i++ {
		s = append(s, &i)
	}
	s[0] = nil
	s = s[1:]
	for _, elem := range s {
		print(*elem)
	}
}

func bad() {
	var s slice
	for i := 0; i < 8; i++ {
		s = append(s, &i)
	}
	s = s[1:]
	for _, elem := range s {
		print(*elem)
	}
}

func goodLog() {
	var log Log
	log.buffer = make(slice, 2)
	log.buffer[0] = nil
	log.buffer = log.buffer[1:]
}

func badLog() {
	var log Log
	log.buffer = make(slice, 2)
	log.buffer = log.buffer[1:]
}

func cutBack() {
	var s slice
	for i := 0; i < 8; i++ {
		s = append(s, &i)
	}
	s[0] = nil
	s = s[:2]
	for _, elem := range s {
		print(*elem)
	}
}

func goodSliceS() {
	var s sliceS
	for i := 0; i < 8; i++ {
		s = append(s, new(S))
	}
	s[0] = nil
	s = s[1:]
	for _, elem := range s {
		print(*elem)
	}
}

func badSliceS() {
	var s sliceS
	for i := 0; i < 8; i++ {
		s = append(s, new(S))
	}
	s = s[1:]
	for _, elem := range s {
		print(*elem)
	}
}
