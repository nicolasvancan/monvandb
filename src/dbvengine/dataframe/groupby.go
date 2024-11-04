package dataframe

import "strings"

var Aggregators = map[string]func(Series) (Element, error){
	"sum":   Sum,
	"count": Count,
	"mean":  Mean,
	"max":   Max,
	"min":   Min,
	"std":   Std,
}

func Sum(s Series) (Element, error) {
	var sum float64
	for _, elem := range s.Elements {
		tmp := elem.ToFloat()

		// Skip null
		if tmp.IsNull() {
			continue
		}

		// Is guaranteed to be a float
		sum += tmp.GetValue().(float64)
	}
	return Elem(sum), nil
}

func Count(s Series) (Element, error) {
	count := 0
	for _, elem := range s.Elements {
		if !elem.IsNull() {
			count++
		}
	}
	return Elem(count), nil
}

func Mean(s Series) (Element, error) {
	var sum float64
	var count int = 0
	for _, elem := range s.Elements {
		tmp := elem.ToFloat()

		// Skip null
		if !tmp.IsNull() {
			// Is guaranteed to be a float
			sum += tmp.GetValue().(float64)
		}

		count++
	}

	return Elem(sum / float64(count)), nil
}

func Max(s Series) (Element, error) {
	var max Element = nil

	for _, elem := range s.Elements {
		if max == nil {
			max = elem
			continue
		}

		if elem.GreaterEqual(max) {
			max = elem
		}
	}
	return max, nil
}

func Min(s Series) (Element, error) {
	var min Element = nil

	for _, elem := range s.Elements {
		if min == nil {
			min = elem
			continue
		}

		if elem.LessEqual(min) {
			min = elem
		}
	}
	return min, nil
}

func Std(s Series) (Element, error) {
	// Calculate the mean
	mean, err := Mean(s)
	if err != nil {
		return nil, err
	}

	// Calculate the variance
	var variance float64
	var count int = 0
	for _, elem := range s.Elements {
		tmp := elem.ToFloat()

		// Skip null
		if !tmp.IsNull() {
			// Is guaranteed to be a float
			variance += (tmp.GetValue().(float64) - mean.ToFloat().GetValue().(float64)) * (tmp.GetValue().(float64) - mean.ToFloat().GetValue().(float64))
		}

		count++
	}

	return Elem(variance / float64(count)), nil
}

// Groupby Aggregator struct
type GroupByAgg struct {
	Column string
	Agg    string
	As     string
}

func NewGroupByAgg(column string, agg string) GroupByAgg {
	return GroupByAgg{
		Column: column,
		Agg:    strings.ToLower(agg),
		As:     "",
	}
}

func createHashTableForColumnsAndAggregators(dataframe Dataframe, columns []string, aggregators []GroupByAgg) ElementHashTable {
	hashTable := make(ElementHashTable)

	goupIndexes, err := getIndexes(dataframe, columns)

	if err != nil {
		return nil
	}

	iterator := NewRowsIterator(&dataframe)

	appendToSlice := func(slice []HashedElementSlice, row HashedElementSlice, indexes []int) []HashedElementSlice {
		elements := make(HashedElementSlice, 0)
		for _, index := range indexes {
			for i, el := range row {
				if i == index {
					elements = append(elements, el)
				}
			}
		}

		slice = append(slice, elements)

		return slice
	}

	for iterator.HasNext() {
		row := iterator.Next()
		hash := createHashForColumns(row, goupIndexes)
		allColumns := make([]string, 0)
		allColumns = append(allColumns, columns...)
		// Append all columns from aggregators
		for _, agg := range aggregators {
			allColumns = append(allColumns, agg.Column)
		}

		// Get all indexes out of the dataframe
		allIndexes, err := getIndexes(dataframe, allColumns)

		if err != nil {
			return nil
		}
		// In case the hash does not exist, create it
		if _, ok := hashTable[hash]; !ok {
			// Create a new row with only groupby columns and aggregators
			hashTable[hash] = make([]HashedElementSlice, 0)
		}

		hashTable[hash] = appendToSlice(hashTable[hash], row, allIndexes)
	}

	return hashTable
}
