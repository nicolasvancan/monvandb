package database

import (
	"bytes"
	"fmt"

	btree "github.com/nicolasvancan/monvandb/src/btree"
	file "github.com/nicolasvancan/monvandb/src/files"
	"github.com/nicolasvancan/monvandb/src/utils"
)

type RangeOptimizerOptions struct {
	// Range Options
	RangeOptions RangeOptions
	// Original Conditions
	Comparsion ColumnComparsion
}

// New range options
func NewRangeOptions() RangeOptions {
	return RangeOptions{
		From:        nil,
		To:          nil,
		Order:       ASC,
		Limit:       -1,
		FComparator: GTE,
		TComparator: GTE,
	}
}

func mergeAnd(r *RangeOptions, other RangeOptions) {
	// Create a new RangeOptions
	if (bytes.Compare(other.From, r.From) < 0 && other.From != nil) || r.From == nil {
		r.From = other.From
	}

	if (bytes.Compare(other.To, r.To) > 0 && other.To != nil) || r.To == nil {
		r.To = other.To
	}
}

func mergeOr(r *RangeOptions, other RangeOptions) {
	// Create a new RangeOptions
	if r.Order == DESC {
		if bytes.Compare(other.From, r.From) < 0 || r.From == nil {
			r.From = other.From
		}

		if bytes.Compare(other.To, r.To) > 0 || r.To == nil {
			r.To = other.To
		}

	} else {
		if bytes.Compare(other.From, r.From) > 0 || r.From == nil {
			r.From = other.From
		}

		if bytes.Compare(other.To, r.To) < 0 || r.To == nil {
			r.To = other.To
		}
	}
}

func (r *RangeOptions) Merge(other RangeOptions, op int) {
	// For end operation we restrict the range based on the lower and maximum values
	if op == AND {
		// Case it is ASC
		mergeAnd(r, other)
	} else {
		mergeOr(r, other)
	}
}

/*
The Range function requires an options for the range. But why is that so? Acctually it's implemented so,] thinking of
possible queries that might reach this point. For instance, let's say that the table has the column "id" as the primary key, thus
the first created data file is indexed by this column. If in one query statement, there is a condition like: WHERE id > 10, the range
query know that it will lookup the 10th key in the index and then start the range from there, running all the dataFile until the end.

If we have a more complex query, like: WHERE id > 10 AND name = "John", the range query will act the same as if it had only the id condition.

Who is going to be responsible for creating the options for the range?

The query parser retrieves all conditions for the query, including joins, and so on. The point is that the lowest end that the system might go
to fetch data is inside the data files, where the functions presented here can be used to retrieve the data.

But what if there is no indexed column in the query? Well, that is the worst case scenario, where the system will scan the whole data file.

The Range function returns a slice of RawRow, and an error if any.

The options for the range are:
- From: The start of the range
- To: The end of the range
- Order: The order of the range, ASC or DESC
- Limit: The limit of the range
- FComparator: The comparator for the From value GTE, GT, EQ, LT, LTE
- TComparator: The comparator for the To value GTE, GT, EQ, LT, LTE

The range query is going to be used in the following queries:
- SELECT * FROM table WHERE id > 10
- SELECT * FROM table WHERE id > 10 AND name = "John"

 Any query that has Where condition with some comparator will use the range query.

 In the future an in memory database will be implemented, it will act as a Proxy, using both Cached data and the data files to
 retrieve needed information.
*/

func Range(t *Table, options RangeOptions) ([]RawRow, error) {
	// If from and to are not set, we return all the rows
	if options.From == nil && options.To == nil && options.Limit < 0 {
		// Scan the whole file
		scannedData, err := scan(options.PDataFile)
		return t.FromKeyValueToRawRow(scannedData), err
	}

	// Get crawler based on the options
	crawler := getCrawlerBasedOnOptions(options)
	// Returns all data gathered from crawling the data file
	return crawlDataFileBasedOnOptions(t, crawler, options)
}

/*
Crawl through the datafile based on the options
*/

func crawlDataFileBasedOnOptions(t *Table, crawler *btree.BTreeCrawler, options RangeOptions) ([]RawRow, error) {
	// Pointer to the function that is used to advance the crawler
	var advance func() error = getCrawlerAdvanceFunction(crawler, options)

	// Loop through the datafile
	limitIdx := 0
	// In case there is a limit
	if options.Limit > -1 {
		rows := make([]RawRow, 0)
		var err error = nil
		for err == nil && limitIdx < options.Limit {
			// Get the key value
			kv := crawler.GetKeyValue()
			// Get the row from the key value
			rows = append(rows, t.FromKeyValueToRawRow(([]btree.BTreeKeyValue{*kv}))[0])

			// Increment the limit index
			limitIdx++
			//Advance
			err = advance()
		}

		return rows, nil
	}

	// Case there is a key to be verified
	if options.To != nil {

		rows := make([]RawRow, 0)
		for advance() == nil {
			// Get the key value
			kv := crawler.GetKeyValue()
			// Get the row from the key value
			rows = append(rows, t.FromKeyValueToRawRow(([]btree.BTreeKeyValue{*kv}))[0])

			// Check the compare to see if we reached the desired results
			if comp, err := compare(kv.Key, options.To, options.TComparator); comp || err != nil {
				break
			}
		}

		return rows, nil
	} else {
		// Where To is nil, we just crawl the data file from begin to end
		rows := make([]RawRow, 0)
		for advance() == nil {
			// Get the key value
			kv := crawler.GetKeyValue()
			// Get the row from the key value
			rows = append(rows, t.FromKeyValueToRawRow(([]btree.BTreeKeyValue{*kv}))[0])
		}
		return rows, nil
	}
}

func getCrawlerAdvanceFunction(crawler *btree.BTreeCrawler, options RangeOptions) func() error {
	var advance func() error
	if options.Order == ASC {
		advance = crawler.Next
	} else {
		advance = crawler.Previous
	}

	return advance
}

/*
Get the crawler based on the options
*/
func getCrawlerBasedOnOptions(options RangeOptions) *btree.BTreeCrawler {
	// Variable to return
	var crawler *btree.BTreeCrawler = nil

	// Change direction based on the order
	if options.Order == ASC {
		crawler = btree.GoToFirstLeaf(options.PDataFile.GetBTree())
	} else {
		crawler = btree.GoToLastLeaf(options.PDataFile.GetBTree())
	}

	// If from is not null, we find the proper key
	if options.From != nil {
		// Find leaf node for the from value
		crawler = options.PDataFile.GetIterator(options.From)
		// Set the crawler position to the from value based on From Comparator
		for {
			comp, err := compare(crawler.GetKeyValue().Key, options.From, options.FComparator)

			if err != nil || !comp {
				crawler.Next()
				continue
			}

			break
			// If the crawler is at the end of the file, break the loop
		}
	}

	return crawler
}

func compare(val1 []byte, val2 []byte, comparator int) (bool, error) {

	switch comparator {
	case GTE:
		return bytes.Compare(val1, val2) >= 0, nil
	case GT:
		return bytes.Compare(val1, val2) > 0, nil
	case LT:
		return bytes.Compare(val1, val2) < 0, nil
	case LTE:
		return bytes.Compare(val1, val2) <= 0, nil
	default:
		break
	}

	return false, fmt.Errorf("invalid comparator")

}

func scan(pDataFile *file.DataFile) ([]btree.BTreeKeyValue, error) {
	// Create a crawler at the beginning of the file
	crawler := btree.GoToFirstLeaf(pDataFile.GetBTree())
	// Loop through the datafile
	keyValues := make([]btree.BTreeKeyValue, 0)

	// Get keyval before entering the loop

	kv := crawler.GetKeyValue()
	if kv == nil {
		return make([]btree.BTreeKeyValue, 0), nil
	}

	keyValues = append(keyValues, *kv)

	for crawler.Next() == nil {
		kv := crawler.GetKeyValue()
		keyValues = append(keyValues, *kv)
	}

	return keyValues, nil
}

/*
Converts []ColumnComparsion to RangeOption fir single comparsion based on simple rules.

The only case that it is possible to infer the exact range of a Range operation is in comparsions where the column is indexed.

Tha happens only for >=, >, <, <= operations. All other cases, even the comparsion with transformation functions, it is hard to infer the range.
One possibility to work also for transformation in to create a inverse function for every function available in the transformation functions.

For instance, if we have a sum function, we can create a inverse function that will be used to infer the range of the comparsion.
But that would cost to much development effort, and that is not the goal right now.


*/

func getRangeOptionsBasedOnColumnComparsion(table *Table, ops ColumnComparsion) RangeOptions {
	// Loop through the operations
	rangeOptions := NewRangeOptions()

	// Consider that there is no indexed column
	rangeOptions.PDataFile = table.PDataFile
	// Check if the column name is indexed
	if table.isColumnIndexed(ops.ColumnName) {
		// If the column name is not the primary key
		if ops.ColumnName != table.PrimaryKey.Name {
			// pointer of DataFile will be passed to rangeOptions pointer
			index := table.Indexes[ops.ColumnName]
			rangeOptions.PDataFile = index.PDataFile
		}
	}

	// Case there is a real value
	if ops.Value.Value != nil {
		// case transformation function is nil
		if ops.Value.Transformation == nil {
			valueBytes, _ := utils.Serialize(ops.Value.Value)

			// Greater
			if ops.Condition == GT {
				// We copy the value to avoid any change in the original value
				rangeOptions.From = valueBytes
				rangeOptions.FComparator = GT
			} else if ops.Condition == GTE {
				rangeOptions.From = valueBytes
				rangeOptions.FComparator = GTE
			} else if ops.Condition == LT {
				rangeOptions.To = valueBytes
				rangeOptions.TComparator = GTE
			} else if ops.Condition == LTE {
				rangeOptions.To = valueBytes
				rangeOptions.TComparator = GT
			}
		}
	}

	return rangeOptions
}

/*
Group operations by indexed columns
*/
func getGroupedOperationsByIndexedColumn(table *Table, ops []ColumnComparsion) map[string][]RangeOptimizerOptions {
	// Loop through the operations
	columnsRangeOptions := make(map[string][]RangeOptimizerOptions)
	// group by columns
	for _, op := range ops {
		// check if the column is indexed
		if table.isColumnIndexed(op.ColumnName) {
			colName := op.ColumnName
			// Case is the first for column, create an empty slice
			if _, ok := columnsRangeOptions[colName]; !ok {
				columnsRangeOptions[colName] = make([]RangeOptimizerOptions, 0)
			}

			columnsRangeOptions[colName] = append(columnsRangeOptions[colName],
				RangeOptimizerOptions{
					RangeOptions: getRangeOptionsBasedOnColumnComparsion(table, op),
					Comparsion:   op,
				})

			continue
		}

		// Bind initial value to pkCol
		pkColName := table.PrimaryKey.Name

		// Verify if it is composed key table
		if table.IsComposedKeyTable() {
			pkColName = table.CompositeKey[0].Name
		}

		// Insert everything in the common pk data file
		if _, ok := columnsRangeOptions[pkColName]; !ok {
			columnsRangeOptions[pkColName] = make([]RangeOptimizerOptions, 0)
		}

		columnsRangeOptions[pkColName] = append(columnsRangeOptions[pkColName],
			RangeOptimizerOptions{
				RangeOptions: getRangeOptionsBasedOnColumnComparsion(table, op),
				Comparsion:   op,
			})
	}

	return columnsRangeOptions
}

func MergeOperationsBasedOnIndexedColumnsAndReturnRangeOptions(table *Table, ops []ColumnComparsion) RangeOptions {
	// Group operations by indexed columns
	groupedOps := getGroupedOperationsByIndexedColumn(table, ops)
	mergedOps := make([]RangeOptimizerOptions, 0)
	for _, ops := range groupedOps {

		lowestLayerOp := findLowestLayerOp(ops)
		mergedOps = append(mergedOps, ops[lowestLayerOp])
	}
	fmt.Printf("MergedOps len %d\n", len(mergedOps))

	preferedRange := 0
	if len(mergedOps) > 1 {
		preferedRange = choosePreferedRange(mergedOps)
	}

	// Returns a full table scan
	if preferedRange == -1 {
		return NewRangeOptions()
	}

	return mergedOps[preferedRange].RangeOptions
}

func findLowestLayerOp(ops []RangeOptimizerOptions) int {
	lowestLayerOp := -1
	for i := 0; i < len(ops)-1; i++ {
		evaluated := i
		if lowestLayerOp == -1 {
			cur := ops[i].Comparsion.LayerLogicalOp
			next := ops[i+1].Comparsion.LayerLogicalOp
			if cur < next {
				lowestLayerOp = i
			} else {
				lowestLayerOp = i + 1
			}
		}

		if lowestLayerOp == i {
			evaluated = i + 1
		}

		op := ops[lowestLayerOp].Comparsion.LayerLogicalOp + ops[evaluated].Comparsion.LayerLogicalOp
		ops[lowestLayerOp].RangeOptions.Merge(ops[evaluated].RangeOptions, op)
	}

	if lowestLayerOp == -1 {
		lowestLayerOp = 0

	}

	return lowestLayerOp
}

func choosePreferedRange(mergedOps []RangeOptimizerOptions) int {
	preferedRange := 0
	fromToNotNullRangesCount := 0 // indicates how many ranges have both from and to values when there is an OR operation

	// Loop through the mergedOps
	for i := 0; i < len(mergedOps)-1; i++ {
		firstCount := countNilValues(mergedOps[i].RangeOptions)
		secondCount := countNilValues(mergedOps[i+1].RangeOptions)

		if mergedOps[i].Comparsion.ParentLogicalOp == OR || mergedOps[i+1].Comparsion.ParentLogicalOp == OR {
			preferedRange = chooseWiderRange(i, i+1, firstCount, secondCount)
			// Count not nil
			fromToNotNullRangesCount += countNotNilRangeBoundaries([]RangeOptions{
				mergedOps[i].RangeOptions,
				mergedOps[i+1].RangeOptions,
			},
			)
		} else {
			preferedRange = chooseNarrowestRange(i, i+1, firstCount, secondCount)
		}
	}

	if fromToNotNullRangesCount > 1 {
		// There is no prefered range, we will return a full table scan
		preferedRange = -1
	}

	return preferedRange
}

func countNilValues(options RangeOptions) int {
	count := 0
	if options.From == nil {
		count++
	}
	if options.To == nil {
		count++
	}
	return count
}

func chooseWiderRange(idx1 int, idx2 int, count1 int, count2 int) int {

	if count1 <= count2 {
		return idx1
	}
	return idx2
}

func chooseNarrowestRange(idx1 int, idx2 int, count1 int, count2 int) int {
	if count1 >= count2 {
		return idx1
	}
	return idx2
}

func countNotNilRangeBoundaries(options []RangeOptions) int {
	count := 0
	for _, opt := range options {
		if opt.From != nil {
			count++
		}
		if opt.To != nil {
			count++
		}
	}
	return count
}

func reverseComparator(comparator int) int {
	switch comparator {
	case GTE:
		return LTE
	case GT:
		return LT
	case LT:
		return GT
	case LTE:
		return GTE
	}

	return comparator
}

func reverseAscToDesc(op *RangeOptions) {
	// Reverse the order
	op.From, op.To = op.To, op.From
	// Reverse the comparators
	op.FComparator = reverseComparator(op.FComparator)
	op.TComparator = reverseComparator(op.TComparator)

}
