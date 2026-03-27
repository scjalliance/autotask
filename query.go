package autotask

// FilterCondition is a single filter expression.
type FilterCondition struct {
	Op    string            `json:"op"`
	Field string            `json:"field,omitempty"`
	UDF   *bool             `json:"udf,omitempty"`
	Value any               `json:"value,omitempty"`
	Items []FilterCondition `json:"items,omitempty"`
}

// FilterQuery is the top-level query structure sent to the API.
type FilterQuery struct {
	Filter []FilterCondition `json:"filter"`
}

// FilterOption is a functional option applied to a FilterQuery, used by query
// methods to compose filter conditions before sending to the API.
type FilterOption func(*FilterQuery)

// Filter creates a FilterQuery from one or more conditions. The returned
// *FilterQuery is directly JSON-marshalable and can also be passed to
// buildFilterQuery via AsFilterOption.
func Filter(conditions ...FilterCondition) *FilterQuery {
	q := &FilterQuery{}
	q.Filter = append(q.Filter, conditions...)
	return q
}

// AsFilterOption converts a *FilterQuery into a FilterOption for use with
// buildFilterQuery.
func (q *FilterQuery) AsFilterOption() FilterOption {
	return func(target *FilterQuery) {
		target.Filter = append(target.Filter, q.Filter...)
	}
}

// buildFilterQuery applies all FilterOptions and returns the combined query.
func buildFilterQuery(opts []FilterOption) *FilterQuery {
	q := &FilterQuery{}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

// FieldSelector holds a field name and whether it is a user-defined field.
type FieldSelector struct {
	name string
	udf  bool
}

// Field returns a FieldSelector for a standard API field.
func Field(name string) FieldSelector {
	return FieldSelector{name: name}
}

// UDFField returns a FieldSelector for a user-defined field.
func UDFField(name string) FieldSelector {
	return FieldSelector{name: name, udf: true}
}

func (f FieldSelector) cond(op string, value ...any) FilterCondition {
	c := FilterCondition{Op: op, Field: f.name}
	if len(value) > 0 {
		c.Value = value[0]
	}
	if f.udf {
		b := true
		c.UDF = &b
	}
	return c
}

func (f FieldSelector) Eq(v any) FilterCondition          { return f.cond("eq", v) }
func (f FieldSelector) NotEq(v any) FilterCondition       { return f.cond("noteq", v) }
func (f FieldSelector) Gt(v any) FilterCondition          { return f.cond("gt", v) }
func (f FieldSelector) Gte(v any) FilterCondition         { return f.cond("gte", v) }
func (f FieldSelector) Lt(v any) FilterCondition          { return f.cond("lt", v) }
func (f FieldSelector) Lte(v any) FilterCondition         { return f.cond("lte", v) }
func (f FieldSelector) BeginsWith(v string) FilterCondition { return f.cond("beginsWith", v) }
func (f FieldSelector) EndsWith(v string) FilterCondition   { return f.cond("endsWith", v) }
func (f FieldSelector) Contains(v string) FilterCondition   { return f.cond("contains", v) }
func (f FieldSelector) Exist() FilterCondition              { return f.cond("exist") }
func (f FieldSelector) NotExist() FilterCondition           { return f.cond("notExist") }
func (f FieldSelector) In(v []any) FilterCondition          { return f.cond("in", v) }

// Or combines multiple conditions with a logical OR.
func Or(conditions ...FilterCondition) FilterCondition {
	return FilterCondition{Op: "or", Items: conditions}
}

// And combines multiple conditions with a logical AND.
func And(conditions ...FilterCondition) FilterCondition {
	return FilterCondition{Op: "and", Items: conditions}
}
