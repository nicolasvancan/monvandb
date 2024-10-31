package dataframe

import (
	"fmt"
)

/*
Filters interface

Filter is an interface for accessing a Tree data structure containing all where conditions.

the tree is built by the parser and the filters are resolved by the dataframe.

The dataframe will iterate over the filters and apply them to the data.

One query example and how it would fit into our system:

SELECT * FROM table1 t1
INNER JOIN table2 t2
WHERE t1.age > 25 AND t2.name = 'Nicolas' OR (t2.age < 30 AND t1.name = 'John')

The parser will build a tree with the following structure:
filters t1
AND

	   |
	  t1.age > 25, OR
					|
					t2.name = 'John'

filters t2

AND

	|

t2.name = 'Nicolas', OR

	|
	t2.age < 30,

The dataframe will iterate over the tree and apply the filters to the data.

Each dataframe has its own filter structure, so the filters are not shared between dataframes.
*/

type FilterNodeType string

const (
	FilterTypeAnd FilterNodeType = "AND"
	FilterTypeOr  FilterNodeType = "OR"
)

type Filters interface {
	Resolve() []ResolvedFilters
	InsertValue(nodeId int, value FilterValue) error
	AddNode(nodeId int, node interface{}, nodeType FilterNodeType) (int, error)
}

type Filter struct {
	Root        *FilterNode
	CurrentNode *FilterNode
	maxID       int
}

type FilterNode struct {
	ID       int
	Type     FilterNodeType
	Value    []FilterValue
	Children []*FilterNode
	Parent   *FilterNode
}

type ResolvedFilters struct {
	Type   FilterNodeType
	Values []FilterValue
}

func (rf *ResolvedFilters) Append(value FilterValue) {
	rf.Values = append(rf.Values, value)
}

type ColumnFilter struct {
	Name       string
	Function   string // Name for the function to be applied
	Parameters []interface{}
}

type FilterValue struct {
	Type       FilterNodeType // AND or OR
	Column     ColumnFilter
	Comparator ComparatorType
	Comparando interface{}
}

func NewFilter() *Filter {
	root := &FilterNode{
		Type:     FilterTypeAnd,
		ID:       0,
		Value:    make([]FilterValue, 0),
		Children: make([]*FilterNode, 0),
	}

	return &Filter{
		Root:        root,
		CurrentNode: root,
		maxID:       0,
	}
}

func nodeToResolvedFilter(node *FilterNode) ResolvedFilters {
	resolved := ResolvedFilters{
		Type:   node.Type,
		Values: node.Value,
	}

	return resolved
}

// This function returns all Node values in a 2D array
// The first dimension is the the order of execution
func (f *Filter) Resolve() []ResolvedFilters {
	resolved := make([]ResolvedFilters, 0)
	resolved = append(resolved, nodeToResolvedFilter(f.CurrentNode))

	for _, node := range f.CurrentNode.Children {
		f.GoToNode(node.ID)
		resolved = append(resolved, f.Resolve()...)
	}
	return resolved
}

func (f *Filter) GoToNode(id int) error {
	// If the id is root, we go to the root node
	if id == f.Root.ID {
		f.CurrentNode = f.Root
		return nil
	}

	for _, node := range f.Root.Children {
		if node.ID == id {
			f.CurrentNode = node
			return nil
		} else {
			newFilter := NewFilter()
			newFilter.Root = node
			newFilter.CurrentNode = node
			newFilter.GoToNode(id)
			if newFilter.CurrentNode.ID == id {
				f.CurrentNode = newFilter.CurrentNode
				return nil
			}

		}
	}

	return fmt.Errorf("node not found")
}

func (f *Filter) AddNode(nodeId int, node interface{}, nodeType FilterNodeType) (int, error) {
	// Go to node
	err := f.GoToNode(nodeId)

	if err != nil {
		return 0, err
	}

	// Add node
	newId := f.maxID + 1

	newNode := &FilterNode{
		ID:       newId,
		Type:     nodeType,
		Value:    make([]FilterValue, 0),
		Children: make([]*FilterNode, 0),
		Parent:   f.CurrentNode,
	}

	f.CurrentNode.Children = append(f.CurrentNode.Children, newNode)
	f.maxID = newId

	f.GoToNode(0)
	return newId, nil
}

func (f *Filter) InsertValue(nodeId int, value FilterValue) error {
	// Go to node
	err := f.GoToNode(nodeId)

	if err != nil {
		return err
	}

	f.CurrentNode.Value = append(f.CurrentNode.Value, value)
	f.GoToNode(0)
	return nil
}
