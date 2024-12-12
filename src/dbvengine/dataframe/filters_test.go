package dataframe

import (
	"testing"
)

func TestFiltersNew(t *testing.T) {
	// Test filters creation
	f := NewFilter()
	if f == nil {
		t.Fatalf("Error creating filters")
	}

	if f.Root == nil {
		t.Errorf("Error creating root node")
	}

	if f.CurrentNode == nil {
		t.Errorf("Error creating current node")
	}

	if f.Root.ID != 0 {
		t.Errorf("Root node ID should not be empty")
	}

	if len(f.Root.Value) != 0 {
		t.Errorf("Root node value should be empty")
	}

	if len(f.Root.Children) != 0 {
		t.Errorf("Root node children should be empty")
	}
}

func TestFiltersGoToNode(t *testing.T) {
	// Test filters go to node
	f := NewFilter()

	f.CurrentNode.Children = append(f.CurrentNode.Children, &FilterNode{
		ID:       1,
		Type:     FilterTypeAnd,
		Value:    []FilterValue{{Column: ColumnFilter{Name: "name"}, Comparator: "eq", Comparando: "John"}},
		Children: make([]*FilterNode, 0),
	})

	err := f.GoToNode(1)

	if err != nil {
		t.Fatalf("Error going to node")
	}

	if f.CurrentNode.ID != 1 {
		t.Errorf("Error getting node")
	}

	err = f.GoToNode(2)

	if err == nil {
		t.Errorf("Error should be thrown")
	}

	f.CurrentNode.Children = append(f.CurrentNode.Children, &FilterNode{
		ID:       2,
		Type:     FilterTypeAnd,
		Value:    []FilterValue{{Column: ColumnFilter{Name: "name"}, Comparator: "eq", Comparando: "John"}},
		Children: make([]*FilterNode, 0),
	})

	err = f.GoToNode(2)

	if err != nil {
		t.Errorf("Error going to node")
	}
}

func TestResolve(t *testing.T) {

	f := NewFilter()

	f.AddNode(0, &FilterNode{
		ID:       1,
		Type:     FilterTypeAnd,
		Value:    []FilterValue{{Type: FilterTypeAnd, Column: ColumnFilter{Name: "name"}, Comparator: "eq", Comparando: "John"}},
		Children: make([]*FilterNode, 0),
	}, FilterTypeAnd)

	f.AddNode(1, &FilterNode{
		ID:       2,
		Type:     FilterTypeAnd,
		Value:    []FilterValue{{Type: FilterTypeAnd, Column: ColumnFilter{Name: "name"}, Comparator: "eq", Comparando: "John"}},
		Children: make([]*FilterNode, 0),
	}, FilterTypeOr)

	resolvedFilters := f.Resolve()

	if resolvedFilters == nil || len(resolvedFilters) != 2 {
		t.Errorf("Error resolving filter")
	}

	if resolvedFilters[0].Type != FilterTypeAnd {
		t.Errorf("Error resolving filter")
	}

	complexFilter := NewFilter()

	complexFilter.InsertValue(0,
		FilterValue{Type: FilterTypeAnd, Column: ColumnFilter{Name: "name"}, Comparator: EQ, Comparando: "John"},
	)

	complexFilter.InsertValue(0,
		FilterValue{Type: FilterTypeAnd, Column: ColumnFilter{Name: "age"}, Comparator: GE, Comparando: 20},
	)

	newNodeValue, err := complexFilter.AddNode(0, &FilterNode{
		Type:     FilterTypeAnd,
		Value:    []FilterValue{{Type: FilterTypeAnd, Column: ColumnFilter{Name: "name"}, Comparator: "eq", Comparando: "Peter"}},
		Children: make([]*FilterNode, 0),
	}, FilterTypeOr)

	if err != nil {
		t.Errorf("Error adding node")
	}

	complexFilter.InsertValue(newNodeValue,
		FilterValue{Type: FilterTypeAnd, Column: ColumnFilter{Name: "age"}, Comparator: GE, Comparando: 20},
	)

	complexFilter.InsertValue(newNodeValue,
		FilterValue{Type: FilterTypeAnd, Column: ColumnFilter{Name: "bang"}, Comparator: BETWEEN, Comparando: []int{20, 30}},
	)

	resolvedFilters = complexFilter.Resolve()

	if resolvedFilters == nil || len(resolvedFilters) != 2 {
		t.Errorf("Error resolving filter")
	}
}
