package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"
)

func main() {
	var i1 Item
	// fmt.Println(i1)
	fmt.Printf("i1: %#v\n", i1)

	i2 := Item{1, 2}
	fmt.Printf("i2: %#v\n", i2)

	i3 := Item{
		Y: 1,
		X: 2,
	}
	fmt.Printf("i3: %#v\n", i3)

	// Can skip when name declaring
	i4 := Item{
		Y: 1,
	}
	fmt.Printf("i4: %#v\n", i4)

	fmt.Println(NewItem(10, 20))
	fmt.Println(NewItem(10, -20))

	i4.Move(100, 200)
	fmt.Printf("i4 (move): %#v\n", i4)

	p1 := Player{
		Name: "P1",
		Item: Item{
			X: 100,
			Y: 200,
		},
	}

	fmt.Printf("p1 with embedded Item: %#v\n", p1)
	fmt.Printf("p1 now has the fields of Item: %#v\n", p1.Item.X)
	fmt.Printf("p1 now has the fields of Item, they are lifted up: %#v\n", p1.X)
	// fmt.Printf("Avoid disambiguity: %#v\n", p1.Item.X)
	fmt.Printf("p1 with uplifted C: %#v\n", p1.C)
	fmt.Printf("p1 with nested C: %#v\n", p1.A.B.C)
	// Methods are uplifted also, since now Player has Item embedded in it, and Item has the Move method
	// INFO: This is NOT inheritance, this is embedding, Move will always get an Item, not a Player
	p1.Move(100, 200)

	ms := []mover{
		&i1,
		// Since Player has Item embedded in it, and Item's methods get uplifted,
		// it means that it implements whatever Item implements too
		&p1,
		&i2,
	}

	moveAll(ms, 0, 0)
	for _, m := range ms {
		fmt.Println(m)
	}

	k := Jade
	fmt.Println("k: ", k)

	// time.Time implements json.Marshaler
	json.NewEncoder(os.Stdout).Encode(time.Now())

	fmt.Println("another key: ", Key(17))

	p1.FoundKey(Jade)
	fmt.Println(string(p1.Keys))
	p1.FoundKey(Jade)
	fmt.Println(string(p1.Keys))

	fmt.Println(p1.Found("copper"))
	fmt.Println(p1.Found("copper"))
	fmt.Println(p1.Found("gold"))
	fmt.Println("keys: ", p1.Keys2)
}

// #EXERCISE:
// - Add a "Keys" field to Player which is a slice of Key
// Add a "FoundKey(k Key) error" method to Player which will add k to Key if it's not there
// 		- Err if k is not one of the known keys

// INFO: This is the method you have to implement on a type to implement how its string representation should be
// Basically Java's `toString()`
// fmt.Stringer interface
func (k Key) String() string {
	switch k {
	case Jade:
		return "Jade"
	case Copper:
		return "Copper"
	case Crystal:
		return "Crystal"
	}

	return fmt.Sprintf("<Key %d>", k)
}

// Go's version of "enum"
const (
	Jade Key = iota + 1
	Copper
	Crystal
	invalidKey // internal, not exported
)

type Key byte

/*
	go >= 1.18
	func NewNumber[T int | float64](kind string) T  {
		if kind == "int" {
			return 0
		}
		return 0.0
	}
*/

// INFO: Rule of thumb, accept interfaces, return types
type mover interface {
	Move(x, y int)
	// Move(int, int)
}

func moveAll(ms []mover, x, y int) {
	for _, m := range ms {
		m.Move(x, y)
	}
}

func (p *Player) FoundKey(k Key) error {
	if k < Jade || k >= invalidKey {
		return fmt.Errorf("invalid Key: %#v", k)
	}

	if !slices.Contains(p.Keys, k) {
		p.Keys = append(p.Keys, k)
	}

	return nil
}

// #EMBEDDING
type Player struct {
	Name string
	Item // Embeds Item
	// T
	A
	Keys  []Key
	Keys2 []string
}

func (p *Player) Found(key string) error {
	switch key {
	case "jade", "copper", "crystal":
		if !slices.Contains(p.Keys2, key) {
			p.Keys2 = append(p.Keys2, key)
		}
		return nil
	default:
		return fmt.Errorf("unknown key - %q", key)
	}
}

// All the fields get lifted
type A struct {
	B
}

type B struct {
	C int
}

// type T struct {
// 	X int
// }

// #METHODS
// `i` is called "the receiver" - basically the "this" keyword in langs like Java
// if you want to mutate, use pointer receiver
func (i *Item) Move(x, y int) {
	i.X = x
	i.Y = y
}

// INFO: Variants of how you can declare a constructor function
// func NewItem(x, y int) Item
// func NewItem(x, y int) *Item
// func NewItem(x, y int) (Item, error)
// func NewItem(x, y int) (*Item, error)
func NewItem(x, y int) (*Item, error) {
	if x < 0 || x > maxX || y < 0 || y > maxY {
		return nil, fmt.Errorf("%d/%d out of bounds %d/%d", x, y, maxX, maxY)
	}

	i := Item{
		X: maxX,
		Y: maxY,
	}

	// The Go compiler does "escape analysis" and will allocate `i` on the heap
	// `go build -gcflags=-m` => to see what is compiler doing
	return &i, nil
}

const (
	maxX = 1_000
	maxY = 600
)

type Item struct {
	X int
	Y int
}
