package filter

// AndFilter to chain logical and
type AndFilter struct {
	filters []UnaryPredicateFilter
}

// And creates instance
func And(filters ...UnaryPredicateFilter) UnaryPredicateFilter {
	cnt := len(filters)
	arr := make([]UnaryPredicateFilter, cnt)

	for i, f := range filters {
		arr[i] = f
	}

	return &AndFilter{
		filters: arr,
	}
}

// Filter implements UnaryPredicateFilter
func (filter *AndFilter) Filter(data []byte) bool {
	for _, f := range filter.filters {
		if !f.Filter(data) {
			return false
		}
	}

	return true
}
