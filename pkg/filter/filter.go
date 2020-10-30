package filter

// UnaryPredicateFilter interface
type UnaryPredicateFilter interface {
	Filter(data []byte) bool
}

// UnaryPredicateFilterFunc for simple filter as a function
type UnaryPredicateFilterFunc func(data []byte) bool

// Filter implements UnaryPredicateFilter
func (f UnaryPredicateFilterFunc) Filter(data []byte) bool {
	return f(data)
}
