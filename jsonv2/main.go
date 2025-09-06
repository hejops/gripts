// https://antonz.org/go-json-v2/

package main

import (
	j1 "encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"encoding/json/jsontext"
	"encoding/json/v2"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	alice := Person{Name: "Alice", Age: 25}
	out := new(strings.Builder) // io.Writer
	out2 := new(strings.Builder)

	// NewEncoder+Encode -> MarshalWrite
	_ = j1.NewEncoder(out).Encode(alice) // NewEncoder removed in v2!
	_ = json.MarshalWrite(out2, alice)

	fmt.Print(out)  // contains trailing newline
	fmt.Print(out2) // no trailing newline
	fmt.Println()

	x := `{"Name":"Bob","Age":30}`
	in := strings.NewReader(x + x) // io.Reader
	var bob Person
	// NewDecoder+Decode -> MarshalRead
	_ = j1.NewDecoder(in).Decode(&bob)
	fmt.Print(bob)
	_ = json.UnmarshalRead(in, &bob)
	fmt.Print(bob)

	fmt.Println(1)

	// people := []Person{{Name: "Alice", Age: 25}, {Name: "Bob", Age: 30}, {Name: "Cindy", Age: 15}}
	// out = new(strings.Builder)
	// enc := jsontext.NewEncoder(out)
	// for _, p := range people {
	// 	_ = json.MarshalEncode(enc, p, json.Deterministic(true) /* opts... */)
	// }
	// fmt.Print(out)

	// jsonl
	in = strings.NewReader(
		` {"Name":"Alice","Age":25} {"Name":"Bob","Age":30} {"Name":"Cindy","Age":15} `,
	)
	dec := jsontext.NewDecoder(in)

	for {
		var p Person
		if err := json.UnmarshalDecode(
			dec,
			&p,
			// &jsonopts.Struct{}, // internal
		); err == io.EOF {
			break
		}
		fmt.Println(p)
	}

	// data, err := json.Marshal(
	// 	true,
	// 	json.WithMarshalers(json.MarshalFunc(func(val bool) ([]byte, error) {
	// 		if val {
	// 			return []byte(`"✓"`), nil
	// 		}
	// 		return []byte(`"✗"`), nil
	// 	})),
	// )
	// fmt.Println(string(data), err)

	var z int
	err := json.Unmarshal(
		[]byte(`3`),
		&z,
		json.WithUnmarshalers(json.UnmarshalFunc(func(b []byte, x *int) error {
			// *x = len(b)
			*x, _ = strconv.Atoi(string(b))
			*x *= 28
			return nil
		})),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(z)

	type Person2 struct {
		Name    string
		Age     *int
		Hobbies []string //`json:"hob,omitempty"`
		Skills  map[string]int
	}

	// // in v1, nil pointers, slices and maps are reduced to `null`, which is
	// // almost never (structurally) useful for clients, thus promoting heavy
	// // use of `omitempty`. however, `omitempty` drops both k and v, which
	// // is also annoying
	// b, _ := j1.Marshal(Person2{Name: "Alice"})
	// fmt.Println(string(b))

	// in v2, nil slices and maps are empty values, which is usually
	// acceptable for clients. nil pointers are still `null` though
	b, _ := json.Marshal(Person2{Name: "Alice"})
	fmt.Println(string(b))

	type Person3 struct {
		FirstName string //`json:"firstname"`
		LastName  string
	}

	var alice3 Person3
	// in v1, in case of duplicate keys, last wins
	j1.Unmarshal([]byte(`{"fIrStNaMe":"Alice","LAStnaMe":"Zakas","firstname":"k"}`), &alice3)
	fmt.Println(alice3) // {k Zakas} -- case-insensitive matching is often surprising
	alice3 = Person3{}
	// in v2, in case of duplicate keys, first wins
	json.Unmarshal([]byte(`{"FirstName":"Alice","lastname":"Zakas","FirstName":"k"}`), &alice3)
	fmt.Println(alice3) // {Alice }

	//
}
