package autotask

import (
	"encoding/json"
	"testing"
)

func TestFilter_SimpleEq(t *testing.T) {
	f := Filter(Field("status").Eq(1))
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"filter":[{"op":"eq","field":"status","value":1}]}`
	if string(data) != want {
		t.Errorf("got %s, want %s", string(data), want)
	}
}

func TestFilter_MultipleConditions_ImplicitAnd(t *testing.T) {
	f := Filter(
		Field("status").Eq(1),
		Field("companyID").Eq(12345),
	)
	data, _ := json.Marshal(f)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	filters := parsed["filter"].([]any)
	if len(filters) != 2 {
		t.Errorf("expected 2 filters, got %d", len(filters))
	}
}

func TestFilter_Or(t *testing.T) {
	f := Filter(
		Or(
			Field("priority").Eq(1),
			Field("priority").Eq(2),
		),
	)
	data, _ := json.Marshal(f)
	want := `{"filter":[{"op":"or","items":[{"op":"eq","field":"priority","value":1},{"op":"eq","field":"priority","value":2}]}]}`
	if string(data) != want {
		t.Errorf("got %s", string(data))
	}
}

func TestFilter_AllOperators(t *testing.T) {
	operators := []struct {
		fn   func() FilterCondition
		want string
	}{
		{func() FilterCondition { return Field("x").Eq(1) }, "eq"},
		{func() FilterCondition { return Field("x").NotEq(1) }, "noteq"},
		{func() FilterCondition { return Field("x").Gt(1) }, "gt"},
		{func() FilterCondition { return Field("x").Gte(1) }, "gte"},
		{func() FilterCondition { return Field("x").Lt(1) }, "lt"},
		{func() FilterCondition { return Field("x").Lte(1) }, "lte"},
		{func() FilterCondition { return Field("x").BeginsWith("a") }, "beginsWith"},
		{func() FilterCondition { return Field("x").EndsWith("z") }, "endsWith"},
		{func() FilterCondition { return Field("x").Contains("mid") }, "contains"},
		{func() FilterCondition { return Field("x").Exist() }, "exist"},
		{func() FilterCondition { return Field("x").NotExist() }, "notExist"},
		{func() FilterCondition { return Field("x").In([]any{1, 2, 3}) }, "in"},
	}
	for _, tt := range operators {
		f := Filter(tt.fn())
		data, _ := json.Marshal(f)
		var parsed map[string]any
		json.Unmarshal(data, &parsed)
		filters := parsed["filter"].([]any)
		cond := filters[0].(map[string]any)
		if cond["op"] != tt.want {
			t.Errorf("expected op=%s, got %s", tt.want, cond["op"])
		}
	}
}

func TestFilter_UDF(t *testing.T) {
	f := Filter(UDFField("CustomerRanking").Eq("Golden"))
	data, _ := json.Marshal(f)
	want := `{"filter":[{"op":"eq","field":"CustomerRanking","udf":true,"value":"Golden"}]}`
	if string(data) != want {
		t.Errorf("got %s", string(data))
	}
}
