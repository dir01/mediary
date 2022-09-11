package service

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAnyToStruct(t *testing.T) {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(`{"people": ["bob", "alice"]}`), &input)
	if err != nil {
		t.Fatal(err)
	}
	type Output struct {
		People []string `json:"people"`
	}
	output := Output{}

	err = mapToStruct(input, &output)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(output.People, []string{"bob", "alice"}) {
		t.Fatalf("expected people to be [bob, alice], got %v", output.People)
	}

}
