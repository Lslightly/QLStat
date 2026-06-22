// Copyright 2026 Qingwei Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
