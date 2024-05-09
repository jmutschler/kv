package kv_test

import (
	"maps"
	"os"
	"testing"

	"github.com/jmutschler/kv"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"kv": kv.Main,
	}))
}

func Test(t *testing.T) {
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

	store, err := kv.OpenStore[string]("testdata/test_countries.kv")
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

func TestOpenStore_ReturnsErrorIfFileNotReadable(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "unreadable.kv"
	err := os.WriteFile(path, nil, 0o000)
	if err != nil {
		t.Fatal(err)
	}
	_, err = kv.OpenStore[string](path)
	if err == nil {
		t.Fatal("want error on opening unreadable file")
	}
}

func TestSetSyncsChangeToDisk(t *testing.T) {
	t.Parallel()
	filename := t.TempDir() + "/countries.kv"
	store, err := kv.OpenStore[string](filename)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Set("GB", "Great Britain")
	if err != nil {
		t.Fatal("Set failed")
	}
	store, err = kv.OpenStore[string](filename)
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
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyncSyncsDataToDisk(t *testing.T) {
	t.Parallel()
	filename := t.TempDir() + "/countries.kv"
	store, err := kv.OpenStore[string](filename)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Set("GB", "Great Britain")
	if err != nil {
		t.Fatal("Set failed")
	}
	err = store.Sync()
	if err != nil {
		t.Fatal(err)
	}
	store, err = kv.OpenStore[string](filename)
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
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetReturnsValueAndOkIfExists(t *testing.T) {
	t.Parallel()

	store, err := kv.OpenStore[string](t.TempDir() + "/countries.kv")
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

	store, err := kv.OpenStore[string](t.TempDir() + "/countries.kv")
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

func TestSyncReturnsErrorForNonWritablePath(t *testing.T) {
	t.Parallel()
	store, err := kv.OpenStore[string](t.TempDir() + "/bogus/data.kv")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Sync()
	if err == nil {
		t.Fatal("want error syncing store to unwritable path")
	}
}

func TestStoreCanBeOfArbitraryType(t *testing.T) {
	t.Parallel()
	store, err := kv.OpenStore[int](t.TempDir() + "/bogus")
	if err != nil {
		t.Fatal(err)
	}
	want := 42
	err = store.Set("answer", want)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := store.Get("answer")
	if !ok {
		t.Fatal("key not found")
	}
	if got != want {
		t.Errorf("want %d, got %d", 42, got)
	}
}
