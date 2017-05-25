package config

import (
	"fmt"
	"reflect"
	"testing"
)

var unmarshalTests = []struct {
	data  string
	value interface{}
}{
	{
		"version: '1.0'",
		&bardConfig{Version: "1.0", Steps: map[string]bardConfigStep{"s": bardConfigStep{"freestyle", "desc"}}},
	},
}

var v = map[string]bardConfigStep{"s": bardConfigStep{"freestyle", "desc"}}
var v2 = &bardConfig{"1.0", map[string]bardConfigStep{"s": bardConfigStep{"freestyle", "desc"}}}

func TestUnmarshal(t *testing.T) {
	for _, item := range unmarshalTests {
		t := reflect.ValueOf(item.value).Type()
		var value interface{}
		switch t.Kind() {
		case reflect.Ptr:
			value = reflect.New(t.Elem()).Interface()
		}
zx		fmt.Println(reflect.ValueOf(item.value).Type().Kind())
	}
	// for _, item := range unmarshalTests {
	// 	t := reflect.ValueOf(item.value).Type()
	// 	var value interface{}
	// 	switch t.Kind() {
	// 	case reflect.Map:
	// 		value = reflect.MakeMap(t).Interface()
	// 	case reflect.String:
	// 		value = reflect.New(t).Interface()
	// 	case reflect.Ptr:
	// 		value = reflect.New(t.Elem()).Interface()
	// 	default:
	// 		//assert.
	// 	}
	// 	err := yaml.Unmarshal([]byte(item.data), value)
	// 	if _, ok := err.(*yaml.TypeError); !ok {
	// 		assert.Equal(t, err, IsNil)
	// 	}
	// 	if t.Kind() == reflect.String {
	// 		assert.Assert(*value.(*string), Equals, item.value)
	// 	} else {
	// 		assert.Assert(value, DeepEquals, item.value)
	// 	}
	// }
}
