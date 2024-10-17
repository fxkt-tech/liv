package json_test

import (
	"fmt"
	"testing"

	"github.com/fxkt-tech/liv/internal/encoding/json"
)

func TestPretty(t *testing.T) {
	x := []byte(`{"x":1}`)
	fmt.Println(json.Pretty(x))
}

type Person struct {
	Name string `json:"name,omitempty"`
	Age  int    `json:"age,omitempty"`
}

func TestToObjectT(t *testing.T) {
	b := []byte(`{"name":"xx","age":1}`)
	x := json.ToV[*Person](b)
	fmt.Println(x.Name, x.Age)
}
