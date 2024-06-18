package database

import (
	"bytes"
	"fmt"

	btree "github.com/nicolasvancan/monvandb/src/btree"
	file "github.com/nicolasvancan/monvandb/src/files"
)

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
			rows = append(rows, t.FromKeyValueToRawRow(([]btree.BTreeKeyValue{kv}))[0])

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
			rows = append(rows, t.FromKeyValueToRawRow(([]btree.BTreeKeyValue{kv}))[0])

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
			rows = append(rows, t.FromKeyValueToRawRow(([]btree.BTreeKeyValue{kv}))[0])
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

	for crawler.Next() == nil {
		kv := crawler.GetKeyValue()
		keyValues = append(keyValues, kv)
	}

	return keyValues, nil
}

/*
Get Range Options based on given TableQueryOperations

This function simply analyses a table and its columns and also the given query operations to return the range options
that will be used in the range query.

*/

func GenerateRangeOptions(t *Table, ops []TableQueryOperation) []RangeOptions {
	// Create a range options
	return nil
}
