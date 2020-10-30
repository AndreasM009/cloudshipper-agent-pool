package filter

// OrFilter to chain logical and
type OrFilter struct {
	filters []UnaryPredicateFilter
}

// Or creates instance
func Or(filters ...UnaryPredicateFilter) UnaryPredicateFilter {
	cnt := len(filters)
	arr := make([]UnaryPredicateFilter, cnt)

	for i, f := range filters {
		arr[i] = f
	}

	return &OrFilter{
		filters: arr,
	}
}

// Filter implements UnaryPredicateFilter
func (filter *OrFilter) Filter(data []byte) bool {
	for _, f := range filter.filters {
		if f.Filter(data) {
			return true
		}
	}

	return false
}
