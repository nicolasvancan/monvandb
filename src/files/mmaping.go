package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

/* Map page */
func MmapPage(f *os.File, page uint64, pageSize uint64) (int, []byte, error) {
	// Get File pointer that comes from os.OpenFile
	fileInfo, err := f.Stat()

	if err != nil {
		return 0, nil, fmt.Errorf("stat %w", err)
	}

	// File is not multiple of pageSize (Must be corrupted or has some bug)
	if fileInfo.Size()%int64(pageSize) != 0 {
		return 0, nil, errors.New("file size may not correspond to database page size")
	}

	// Get mmap based on given page
	pageData, err := syscall.Mmap(
		int(f.Fd()),
		int64(page*pageSize),
		int(pageSize),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED)

	if err != nil {
		return 0, nil, fmt.Errorf("mmap Error %w", err)
	}

	return int(fileInfo.Size()), pageData, nil
}
