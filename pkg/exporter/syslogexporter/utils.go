package syslogexporter

import "fmt"

type errorWithCount struct {
	err   error
	count int
}

// deduplicateErrors replaces duplicate instances of the same error in a slice
// with a single error containing the number of times it occurred added as a suffix.
// For example, three occurrences of "error: 502 Bad Gateway"
// are replaced with a single instance of "error: 502 Bad Gateway (x3)".
func deduplicateErrors(errs []error) []error {
	if len(errs) < 2 {
		return errs
	}
	errorsWithCounts := []errorWithCount{}
	for _, err := range errs {
		found := false
		for i := range errorsWithCounts {
			if errorsWithCounts[i].err.Error() == err.Error() {
				found = true
				errorsWithCounts[i].count += 1
				break
			}
		}
		if !found {
			errorsWithCounts = append(errorsWithCounts, errorWithCount{
				err:   err,
				count: 1,
			})
		}
	}
	var uniqueErrors []error
	for _, errorWithCount := range errorsWithCounts {
		if errorWithCount.count == 1 {
			uniqueErrors = append(uniqueErrors, errorWithCount.err)
		} else {
			uniqueErrors = append(uniqueErrors, fmt.Errorf("%s (x%d)", errorWithCount.err, errorWithCount.count))
		}
	}
	return uniqueErrors
}

func errorListToStringSlice(errList []error) []string {
	errStrList := make([]string, len(errList))
	for i, err := range errList {
		errStrList[i] = err.Error()
	}
	return errStrList
}
