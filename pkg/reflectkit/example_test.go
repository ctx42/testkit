package reflectkit

import (
	"fmt"
	"testing"
)

func ExampleGetField() {
	type Event struct {
		ID    string `json:"id" validate:"required"`
		Score int    `json:"score"`
	}
	e := &Event{}
	fld := GetField(&testing.T{}, e, "ID")
	fmt.Println(fld.Tag.Get("json"))
	fmt.Println(fld.Tag.Get("validate"))
	// Output:
	// id
	// required
}

func ExampleGetValue() {
	type Event struct {
		ID    string `json:"id" validate:"required"`
		Score int    `json:"score"`
	}
	e := &Event{ID: "evt-1"}
	val := GetValue(&testing.T{}, e, "ID")
	fmt.Println(val.Kind())
	fmt.Println(val.String())
	// Output:
	// string
	// evt-1
}
