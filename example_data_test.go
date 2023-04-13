package skipper_test

import (
	"fmt"

	"github.com/lukasjarosch/skipper"
)

func ExampleNewData_struct() {
	type MyType struct {
		Foo   string
		Items []string
	}
	myType := MyType{
		Foo:   "Hello World",
		Items: []string{"Apple", "Banana", "Pie"},
	}

	// creating [Data] from a custom struct type
	data, _ := skipper.NewData(myType)
	fmt.Println(data)

	// Output:
	// map[foo:Hello World items:[Apple Banana Pie]]
}

func ExampleNewData_map() {
	exampleMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": "baz",
			"qux": []interface{}{
				"hello",
				"world",
			},
		},
	}
	// Here we're simulating a simple yaml/json file by passing in a map[string]interface{}.
	dataMap, _ := skipper.NewData(exampleMap)
	fmt.Println(dataMap)

	// Output:
	// map[foo:map[bar:baz qux:[hello world]]]
}

func ExampleData_GetPath() {
	exampleData := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": "baz",
			"qux": []interface{}{
				"hello",
				"world",
			},
			"baz": []interface{}{
				map[string]interface{}{
					"apple": "healthy",
				},
				map[string]interface{}{
					"pizza": "yum",
				},
			},
		},
	}
	data, _ := skipper.NewData(exampleData)

	// accessing a map works by just building the path by appending map keys
	baz, _ := data.GetPath(skipper.P("foo.bar"))

	// if the data contains an array, simply add the index of the element to access
	world, _ := data.GetPath(skipper.P("foo.qux.1"))

	// if the array is a list of maps, you can further traverse into the element just as you'd expect
	yum, _ := data.GetPath(skipper.P("foo.baz.1.pizza"))

	fmt.Println(baz)
	fmt.Println(world)
	fmt.Println(yum)

	// Output:
	// baz
	// world
	// yum
}

func ExampleData_UnmarshalPath() {
	exampleData := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": "baz",
			"qux": []interface{}{
				"hello",
				"world",
			},
		},
	}
	data, _ := skipper.NewData(exampleData)

	type MyType struct {
		Words []string `yaml:"qux"`
	}

	myType := new(MyType)

	data.UnmarshalPath(skipper.P("foo"), myType)

	fmt.Printf("%+v", myType)

	// Output:
	// &{Words:[hello world]}
}
