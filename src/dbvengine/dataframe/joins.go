package dataframe

import "fmt"

type HashedElementSlice []Element
type ElementHashTable map[string][]HashedElementSlice

func isColumnInList(column string, list []string) bool {
	for _, col := range list {
		if col == column {
			return true
		}
	}
	return false
}

// simply concat columns values of a row
func createHashForColumns(row []Element, indexes []int) string {
	hash := ""
	for _, index := range indexes {
		hash += row[index].String() + "|"
	}
	return hash
}

func getIndexes(df Dataframe, on []string) ([]int, error) {
	indexes := make([]int, len(on))
	for i, on := range on {
		tmpIndex := indexOf(on, df.Columns)

		if tmpIndex == -1 {
			return nil, fmt.Errorf("column %s not found in df1", on)
		}

		indexes[i] = tmpIndex
	}
	return indexes, nil
}

func createHashTableForDf(dataframe Dataframe, on []string) (ElementHashTable, error) {
	hashTable := make(ElementHashTable)
	indexes, err := getIndexes(dataframe, on)

	if err != nil {
		return nil, err
	}

	iterator := NewRowsIterator(&dataframe)

	for iterator.HasNext() {
		row := iterator.Next()
		addRowToHash(row, &hashTable, indexes)
	}

	return hashTable, nil
}

func addRowToHash(row HashedElementSlice, hashTable *ElementHashTable, indexes []int) {
	hash := createHashForColumns(row, indexes)

	if _, ok := (*hashTable)[hash]; !ok {
		(*hashTable)[hash] = make([]HashedElementSlice, 0)
	}

	(*hashTable)[hash] = append((*hashTable)[hash], row)
}

func resolveColumnsAndDTypesForJoin(df1 Dataframe, df2 Dataframe, on []string) ([]string, []int, error) {
	// Inner columns are supposed to be all columns
	// without duplicating the ones in on list

	cols := make([]string, 0)
	dtypes := make([]int, 0)

	for _, col := range df1.Columns {
		cols = append(cols, col.Info.Name)
		dtypes = append(dtypes, col.DType)

	}

	for i, col := range df2.Columns {
		if !isColumnInList(col.Info.Name, on) {
			if isColumnInList(col.Info.Name, cols) {
				if df2.Alias != "" {
					cols = append(cols, df2.Alias+"."+col.Info.Name)
					dtypes = append(dtypes, df2.Columns[i].DType)
				} else {
					return nil, nil, fmt.Errorf("column %s exists in both dataframes", col.Info.Name)
				}
			} else {
				cols = append(cols, col.Info.Name)
				dtypes = append(dtypes, df2.Columns[i].DType)
			}
		}
	}

	return cols, dtypes, nil
}

func inIndex(index int, indexes []int) bool {
	for _, indx := range indexes {
		if indx == index {
			return true
		}
	}
	return false
}

func innerJoin(df1 Dataframe, df2 Dataframe, on []string) (Dataframe, error) {

	hashTable, err := createHashTableForDf(df1, on)

	if err != nil {
		return Dataframe{}, err
	}

	// For dataFrame 2
	indexes, err := getIndexes(df2, on)

	if err != nil {
		return Dataframe{}, err
	}

	iterator := NewRowsIterator(&df2)

	// Resolve columns and dTypes
	resolvedColumns, dTypes, err := resolveColumnsAndDTypesForJoin(df1, df2, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Create a new Dataframe
	finalDf := NewDataframe(resolvedColumns)
	finalDf.setDtypes(dTypes)

	for iterator.HasNext() {
		row := iterator.Next()
		hash := createHashForColumns(row, indexes)

		hashedValues, ok := hashTable[hash]
		// If the hash exists in the hash table
		if ok {

			for _, hashedValue := range hashedValues {
				colsToAppend := make([]Element, 0)

				for i, el := range row {
					if !inIndex(i, indexes) {
						colsToAppend = append(colsToAppend, el)
					}
				}

				finalDf.AddRow(append(hashedValue, colsToAppend...))
			}
		}
	}

	return finalDf, nil
}

func leftJoin(df1 Dataframe, df2 Dataframe, on []string) (Dataframe, error) {

	hashTableDf1, err := createHashTableForDf(df1, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Resolve columns and dTypes
	resolvedColumns, dTypes, err := resolveColumnsAndDTypesForJoin(df1, df2, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Create a new Dataframe
	finalDf := NewDataframe(resolvedColumns)
	finalDf.setDtypes(dTypes)

	hasTableDf2, err := createHashTableForDf(df2, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Indexes For dataFrame 2
	indexes, err := getIndexes(df2, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Iterate over the first hashtable
	for hash, hashedValues1 := range hashTableDf1 {

		// Iterate over all hashed elements for the given hash
		for _, elements := range hashedValues1 {

			// Get hash value []Elements
			hashedValues, ok := hasTableDf2[hash]

			// If the hash exists in the right hash table
			if ok {
				for _, hashedValue := range hashedValues {
					for i, el := range hashedValue {
						if !inIndex(i, indexes) {
							elements = append(elements, el)
						}
					}
					finalDf.AddRow(elements)
				}

			} else {

				for len(resolvedColumns)-len(elements) > 0 {
					elements = append(elements, Elem(nil))
				}
				finalDf.AddRow(elements)
			}

		}
	}

	// Create a new Dataframe
	return finalDf, nil
}

func rightJoin(df1 Dataframe, df2 Dataframe, on []string) (Dataframe, error) {
	// Operates exactly like left join but with the dataframes switched
	return leftJoin(df2, df1, on)
}

func outerJoin(df1 Dataframe, df2 Dataframe, on []string) (Dataframe, error) {
	// Create a new Dataframe
	hashTableDf1, err := createHashTableForDf(df1, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Resolve columns and dTypes
	resolvedColumns, dTypes, err := resolveColumnsAndDTypesForJoin(df1, df2, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Create a new Dataframe
	finalDf := NewDataframe(resolvedColumns)
	finalDf.setDtypes(dTypes)

	hasTableDf2, err := createHashTableForDf(df2, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Indexes For dataFrame 2
	indexes, err := getIndexes(df2, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Iterate over the first hashtable
	for hash, hashedValues1 := range hashTableDf1 {

		// Iterate over all hashed elements for the given hash
		for _, elements := range hashedValues1 {

			// Get hash value []Elements
			hashedValues, ok := hasTableDf2[hash]

			// If the hash exists in the right hash table
			if ok {
				for _, hashedValue := range hashedValues {
					for i, el := range hashedValue {
						if !inIndex(i, indexes) {
							elements = append(elements, el)
						}
					}
					finalDf.AddRow(elements)
				}

			} else {

				for len(resolvedColumns)-len(elements) > 0 {
					elements = append(elements, Elem(nil))
				}
				finalDf.AddRow(elements)
			}

		}
	}

	indexesDf1, err := getIndexes(df1, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Iterate over the second hashtable
	for hash, hashedValues2 := range hasTableDf2 {
		// Set all elements of the hash table that are not in index to nil
		// To get that we need to know what are those elements
		//
		// Since the left side of the operation is df1, its columns should be set first
		// Exactly in resolved columns order
		//
		// Therefore, we create a slice of the length of the df1 dataframe

		df1Elements := make([]Element, 0)
		for _, elements := range hashedValues2 {

			if _, ok := hashTableDf1[hash]; !ok {

				// Fill it up with nulls
				for len(df1Elements) < len(df1.Columns) {
					df1Elements = append(df1Elements, Elem(nil))
				}

				// The second part is to find which columns of df1 are join columns and also in df2
				// Knowing that, we can set the values of those columns to the values of the hash table

				for i := range df1Elements {
					if inIndex(i, indexesDf1) {
						indexColumnLocation := indexOf(df1.Columns[i].Info.Name, df2.Columns)
						indexLocation := 0
						for j, colIdx := range indexes {
							if colIdx == indexColumnLocation {
								indexLocation = j
								break
							}
						}
						df1Elements[i] = elements[indexes[indexLocation]]
					}
				}

				for i, el := range elements {
					if !inIndex(i, indexes) {
						df1Elements = append(df1Elements, el)
					}
				}
				finalDf.AddRow(df1Elements)
			}

		}
	}

	return finalDf, nil
}
