package gotmpl

import "testing"

func TestMapLookupTemplate(t *testing.T) {
	s := "echo ${foo}"
	l := map[string]string{"foo": "bar"}

	res, err := TemplateString(s, MapLookup(l))
	if err != nil {
		t.Fatal(err)
	}
	if res != "echo bar" {
		t.Errorf("Expected `echo bar`, got %v", res)
	}
}

func TestEarlyEOFTemplate(t *testing.T) {
	_, err := TemplateString(`echo ${foo`, MapLookup(map[string]string{"foo": "bar"}))
	if err == nil {
		t.Error("Should fail on mismatched bracse")
	}
}
