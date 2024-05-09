package kv

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Store[T any] struct {
	path string
	data map[string]T
}

type Args struct {
	Path  string
	Key   string
	Value string
	Verb  string
}

func OpenStore[T any](path string) (*Store[T], error) {
	s := &Store[T]{
		path: path,
		data: map[string]T{},
	}

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&s.data)
	if err != nil {
		return nil, fmt.Errorf("could not decode %v: %v", path, err)
	}

	return s, nil
}

func (s Store[T]) Get(key string) (T, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s *Store[T]) Set(key string, value T) error {
	s.data[key] = value
	return s.Sync()
}

func (s *Store[T]) Sync() error {
	file, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(s.data)
	if err != nil {
		return fmt.Errorf("could not encode %v: %v", s.path, err)
	}

	return nil
}

func (s Store[T]) All() map[string]T {
	return s.data
}

func (s *Store[T]) Close() error {
	return s.Sync()
}

func Main() int {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: kv <path> [key] [value]")
		return 1
	}

	args, _ := ParseArgs(os.Args[1:])

	store, err := OpenStore[string](args.Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	switch args.Verb {
	case "list":
		data := store.All()

		for k, v := range data {
			fmt.Printf("%s:%s\n", k, v)
		}

	case "get":
		v, ok := store.Get(args.Key)
		if !ok {
			fmt.Fprintln(os.Stderr, "Key not found")
			return 1
		}
		fmt.Println(v)

	default:
		err := store.Set(args.Key, args.Value)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	return 0
}

func ParseArgs(args []string) (Args, error) {
	a := Args{Path: "default.kv"}

	if len(args) > 0 {
		if strings.HasSuffix(args[0], ".kv") {
			a.Path = args[0]
			args = args[1:]
		}
	}

	if len(args) == 0 {
		a.Verb = "list"
		return a, nil
	} else if len(args) == 1 {
		a.Verb = "get"
		a.Key = args[0]
		return a, nil
	}

	a.Verb = "set"
	a.Key = args[0]
	a.Value = strings.Join(args[1:], " ")

	return a, nil
}
