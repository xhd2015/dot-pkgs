package space

import (
	"fmt"
	"sort"
)

// MockBackend is an in-memory Backend for tests.
type MockBackend struct {
	Desktops   []int
	Created    int
	Switched   []int
	FailCreate error
	FailSwitch error
	FailList   error
}

func (m *MockBackend) Create() error {
	if m.FailCreate != nil {
		return m.FailCreate
	}
	m.Created++
	h := m.highest()
	if h < 1 {
		h = 0
	}
	m.Desktops = append(m.Desktops, h+1)
	return nil
}

func (m *MockBackend) Switch(n int) error {
	if m.FailSwitch != nil {
		return m.FailSwitch
	}
	if n < 1 {
		return fmt.Errorf("invalid desktop number: %d", n)
	}
	found := false
	for _, d := range m.Desktops {
		if d == n {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("desktop not found: Desktop %d", n)
	}
	m.Switched = append(m.Switched, n)
	return nil
}

func (m *MockBackend) List() ([]Desktop, error) {
	if m.FailList != nil {
		return nil, m.FailList
	}
	nums := append([]int(nil), m.Desktops...)
	sort.Ints(nums)
	out := make([]Desktop, 0, len(nums))
	for _, n := range nums {
		out = append(out, Desktop{Number: n, Name: fmt.Sprintf("Desktop %d", n)})
	}
	return out, nil
}

func (m *MockBackend) Highest() (int, error) {
	h := m.highest()
	if h < 1 {
		return 0, fmt.Errorf("no Desktop buttons found")
	}
	return h, nil
}

func (m *MockBackend) highest() int {
	h := 0
	for _, d := range m.Desktops {
		if d > h {
			h = d
		}
	}
	return h
}
