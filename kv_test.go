package kv_test

import (
	"kv"
	"maps"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"kv": kv.Main,
	}))
}

func TestKVWithPathNameArgPrintsContentsOfStore(t *testing.T) {

	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

func TestKVParseArgsReturnsList(t *testing.T) {

	tests := []struct {
		args []string
		want kv.Args
	}{
		{
			args: []string{"store.kv"},
			want: kv.Args{Path: "store.kv", Verb: "list"},
		},
		{
			args: []string{},
			want: kv.Args{Path: "default.kv", Verb: "list"},
		},
	}

	for _, tt := range tests {

		got, err := kv.ParseArgs(tt.args)
		if err != nil {
			t.Fatal(err)
		}

		if tt.want != got {
			t.Fatalf("want %+v, got %+v", tt.want, got)
		}
	}

}

func TestKVParseArgsReturnsGet(t *testing.T) {

	tests := []struct {
		args []string
		want kv.Args
	}{
		{
			args: []string{"store.kv", "us"},
			want: kv.Args{Path: "store.kv", Verb: "get", Key: "us"},
		},
		{
			args: []string{"us"},
			want: kv.Args{Path: "default.kv", Verb: "get", Key: "us"},
		},
	}

	for _, tt := range tests {

		got, err := kv.ParseArgs(tt.args)
		if err != nil {
			t.Fatal(err)
		}

		if tt.want != got {
			t.Fatalf("want %+v, got %+v", tt.want, got)
		}
	}
}

func TestKVParseArgsReturnsSet(t *testing.T) {

	tests := []struct {
		args []string
		want kv.Args
	}{
		{
			args: []string{"store.kv", "us", "united", "states"},
			want: kv.Args{Path: "store.kv", Verb: "set", Key: "us", Value: "united states"},
		},
		{
			args: []string{"us", "pizza", "is", "good"},
			want: kv.Args{Path: "default.kv", Verb: "set", Key: "us", Value: "pizza is good"},
		},
	}

	for _, tt := range tests {

		got, err := kv.ParseArgs(tt.args)
		if err != nil {
			t.Fatal(err)
		}

		if tt.want != got {
			t.Fatalf("want %+v, got %+v", tt.want, got)
		}
	}
}

func TestOpenStore(t *testing.T) {
	t.Parallel()

	store, err := kv.OpenStore("testdata/test_countries.kv")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"US": "United States",
		"GB": "Great Britain",
		"CA": "Canada",
	}

	got := store.All()

	if !maps.Equal(want, got) {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestDataPersistsAfterClose(t *testing.T) {
	t.Parallel()

	filename := t.TempDir() + "/countries.kv"

	store, err := kv.OpenStore(filename)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Set("GB", "Great Britain")
	if err != nil {
		t.Fatal("Set failed")
	}

	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}

	store, err = kv.OpenStore(filename)
	if err != nil {
		t.Fatal(err)
	}

	got, ok := store.Get("GB")
	if !ok {
		t.Fatal("key not found")
	}
	if got != "Great Britain" {
		t.Fatal("wrong value")
	}
}

func TestGetReturnsValueAndOkIfExists(t *testing.T) {
	t.Parallel()

	store, err := kv.OpenStore(t.TempDir() + "/countries.kv")
	if err != nil {
		t.Fatal(err)
	}

	want := "Great Britain"

	err = store.Set("GB", want)
	if err != nil {
		t.Fatal("Set failed")
	}

	got, ok := store.Get("GB")
	if !ok {
		t.Fatal("want ok = true, got ok = false")
	}
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}

}

func TestGetReturnsEmptyValueAndNotOkIfNotExists(t *testing.T) {
	t.Parallel()

	store, err := kv.OpenStore(t.TempDir() + "/countries.kv")
	if err != nil {
		t.Fatal(err)
	}

	want := ""

	got, ok := store.Get("keythatdoesn'texist")
	if ok {
		t.Fatal("want ok = false, got ok = true")
	}
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}
