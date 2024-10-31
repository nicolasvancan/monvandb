package dataframe

import "fmt"

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

func createHashTableForDf(dataframe Dataframe, on []string) (map[string][]Element, error) {
	hashTable := make(map[string][]Element)
	indexes, err := getIndexes(dataframe, on)

	if err != nil {
		return nil, err
	}

	iterator := NewRowsIterator(&dataframe)

	for iterator.HasNext() {
		row := iterator.Next()
		hash := createHashForColumns(row, indexes)
		hashTable[hash] = row
	}

	return hashTable, nil
}

func resolveColumnsAndDTypesForJoin(df1 Dataframe, df2 Dataframe, on []string) ([]string, []int, error) {
	// Inner columns are supposed to be all columns
	// without duplicating the ones in on list

	cols := make([]string, 0)
	cols = append(cols, df1.Columns...)
	dtypes := make([]int, 0)
	dtypes = append(dtypes, df1.DTypes...)

	for i, col := range df2.Columns {
		if !isColumnInList(col, on) {
			if isColumnInList(col, cols) {
				if df2.Alias != "" {
					cols = append(cols, df2.Alias+"."+col)
					dtypes = append(dtypes, df2.DTypes[i])
				} else {
					return nil, nil, fmt.Errorf("column %s exists in both dataframes", col)
				}
			} else {
				cols = append(cols, col)
				dtypes = append(dtypes, df2.DTypes[i])
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
		if _, ok := hashTable[hash]; ok {
			colsToAppend := make([]Element, 0)

			for i, el := range row {
				if !inIndex(i, indexes) {
					colsToAppend = append(colsToAppend, el)
				}
			}

			finalDf.AddRow(append(hashTable[hash], colsToAppend...))
			finalDf.nRows++
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
	for hash, elements := range hashTableDf1 {

		if _, ok := hasTableDf2[hash]; ok {
			hashValue := hasTableDf2[hash]
			for i, el := range hashValue {
				if !inIndex(i, indexes) {
					elements = append(elements, el)
				}
			}
		} else {
			for len(resolvedColumns)-len(elements) > 0 {
				elements = append(elements, Elem(nil))
			}
		}

		finalDf.AddRow(elements)
		finalDf.nRows++
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
	for hash, elements := range hashTableDf1 {

		if _, ok := hasTableDf2[hash]; ok {
			// Get hash value []Elements
			hashValue := hasTableDf2[hash]

			for i, el := range hashValue {
				if !inIndex(i, indexes) {
					elements = append(elements, el)
				}
			}
		} else {
			for len(resolvedColumns)-len(elements) > 0 {
				elements = append(elements, Elem(nil))
			}
		}

		finalDf.AddRow(elements)
		finalDf.nRows++
	}

	indexesDf1, err := getIndexes(df1, on)

	if err != nil {
		return Dataframe{}, err
	}

	// Iterate over the second hashtable
	for hash, elements := range hasTableDf2 {
		// Set all elements of the hash table that are not in index to nil
		// To get that we need to know what are those elements
		//
		// Since the left side of the operation is df1, its columns should be set first
		// Exactly in resolved columns order
		//
		// Therefore, we create a slice of the length of the df1 dataframe

		df1Elements := make([]Element, 0)

		if _, ok := hashTableDf1[hash]; !ok {

			// Fill it up with nulls
			for len(df1Elements) < len(df1.Columns) {
				df1Elements = append(df1Elements, Elem(nil))
			}

			// The second part is to find which columns of df1 are join columns and also in df2
			// Knowing that, we can set the values of those columns to the values of the hash table

			for i := range df1Elements {
				if inIndex(i, indexesDf1) {
					indexLocation := indexOf(df1.Columns[i], df2.Columns)
					df1Elements[i] = elements[indexes[indexLocation]]
				}
			}

			for i, el := range elements {
				if !inIndex(i, indexes) {
					df1Elements = append(df1Elements, el)
				}
			}
			finalDf.AddRow(df1Elements)
			finalDf.nRows++
		}

	}

	// Create a new Dataframe
	return finalDf, nil
}
