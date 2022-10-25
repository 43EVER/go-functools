package gofunctools_test

import (
	"net/url"
	"reflect"
	"testing"

	gofunctools "github.com/43EVER/go-functools"
)

func TestBindFirst(t *testing.T) {
	f := func(a int32) int32 { return a + 1 }
	f2 := gofunctools.BindFirst(f, 1)
	if f2() != 2 {
		t.Errorf("expected 2, but got %d", f2())
		return
	}
}

func TestBind2First(t *testing.T) {
	f := func(a int32, b int32) int32 { return a + b }
	f2 := gofunctools.Bind2First(f, 1)
	if f2(6) != 7 {
		t.Errorf("expected 7, but got %d", f2(6))
		return
	}
}

func TestReverse(t *testing.T) {
	f := func(a string, b string) string {
		return a + b
	}
	f2 := gofunctools.ReverseParamter(f)
	if f2("456", "123") != "123456" {
		t.Errorf("expected 123456, but got %s", f2("456", "123"))
	}
}

func TestStructToMap(t *testing.T) {
	s := struct {
		FirstName string
		LastName  string
		Age       int32
		Map       map[string]struct {
			Key1 string
			Key2 string
		}
		Struct *struct {
			SKey1 string
			SKey2 string
		}
	}{
		FirstName: "jiaran",
		LastName:  "wang",
		Age:       23,
		Map: map[string]struct {
			Key1 string
			Key2 string
		}{
			"1": {Key1: "value1", Key2: "value2"},
			"2": {Key1: "value1_", Key2: "value2_"},
		},
		Struct: &struct {
			SKey1 string
			SKey2 string
		}{
			SKey1: "skey1_value1",
			SKey2: "skey2_value2",
		},
	}

	result := gofunctools.StructToMap(s)

	if result["FirstName"].(string) != s.FirstName ||
		result["LastName"].(string) != s.LastName ||
		result["Age"].(int32) != s.Age {
		t.Error("data of struct checks failed after converting.")
		return
	}
	t.Logf("%#v", result)
}

func TestUnique(t *testing.T) {
	data := []int{1, 1, 2, 3, 3, 4}
	uniqueData := gofunctools.Unique(data)
	if len(uniqueData) != 4 {
		t.Errorf("count of elements if less than 4")
	}

	for idx1, v1 := range uniqueData {
		for idx2, v2 := range uniqueData {
			if idx1 != idx2 && v1 == v2 {
				t.Errorf("repeated, idx1:%d, idx2:%d value:%d", idx1, idx2, v1)
			}
		}
	}

	data = []int{}
	uniqueData = data
	if len(uniqueData) != 0 {
		t.Errorf("want [], got %v", uniqueData)
	}
}

func TestAndSlice(t *testing.T) {
	a := []int{1, 1, 3, 3, 2, 2}
	b := []int{3, 2, 2, 5}

	want := []int{3, 2}
	got := gofunctools.AndSlice(a, b)
	if len(want) != len(got) {
		t.Errorf("want %d items slice, got %d", len(want), len(got))
		return
	}

	for idx, wantItem := range want {
		if wantItem != got[idx] {
			t.Errorf("%d th item should be %d, got %d", idx, want, got)
			return
		}
	}
}

func TestSubSlice(t *testing.T) {
	a := []int{1, 1, 3, 3, 2, 2}
	b := []int{3, 2, 2, 5}

	want := []int{1}
	got := gofunctools.SubSlice(a, b)
	if len(want) != len(got) {
		t.Errorf("want %d items slice, got %d", len(want), len(got))
		return
	}

	for idx, wantItem := range want {
		if wantItem != got[idx] {
			t.Errorf("%d th item should be %d, got %d", idx, want, got)
			return
		}
	}
}

func TestGenGetFieldFunc(t *testing.T) {
	// struct
	type structT = struct {
		I int64
		S string
	}
	a := structT{
		I: 123,
		S: "str",
	}
	fun := gofunctools.GenGetFieldFunc[structT, int64]("I")
	if a.I != fun(a) {
		t.Errorf("want %d, got %d", a.I, fun(a))
		t.FailNow()
	}

	// pointer to pointer of struct
	pointerFun := gofunctools.GenGetFieldFunc[**structT, string]("S")
	pointerOfA := &a
	if a.S != pointerFun(&pointerOfA) {
		t.Errorf("want %d, got %d", a.I, fun(a))
		t.FailNow()
	}

	// map
	mapData := map[string]interface{}{
		"123": 123,
	}
	mapFun := gofunctools.GenGetFieldFunc[map[string]interface{}, interface{}]("123")
	if mapData["123"] != mapFun(mapData) {
		t.Errorf("want %d, got %d", a.I, fun(a))
		t.FailNow()
	}

	//TODO 数组和 struct 和 map 嵌套，暂时不需要，有需求再说
}

func TestZip(t *testing.T) {
	type args struct {
		s1 []int
		s2 []string
	}
	tests := []struct {
		name string
		args args
		want []gofunctools.Tuple[int, string]
	}{
		{
			"正常",
			args{
				[]int{1, 2, 3},
				[]string{"a", "b", "c"},
			},
			[]gofunctools.Tuple[int, string]{
				{1, "a"},
				{2, "b"},
				{3, "c"},
			},
		},
		{
			"1比2长",
			args{
				[]int{1, 2, 3},
				[]string{"a", "b"},
			},
			[]gofunctools.Tuple[int, string]{
				{1, "a"},
				{2, "b"},
			},
		},
		{
			"1比2短",
			args{
				[]int{1, 2},
				[]string{"a", "b", "c"},
			},
			[]gofunctools.Tuple[int, string]{
				{1, "a"},
				{2, "b"},
			},
		},
		{
			"1为空",
			args{
				[]int{},
				[]string{"a", "b", "c"},
			},
			[]gofunctools.Tuple[int, string]{},
		},
		{
			"2为空",
			args{
				[]int{1, 2},
				[]string{},
			},
			[]gofunctools.Tuple[int, string]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gofunctools.Zip(tt.args.s1, tt.args.s2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Zip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMax(t *testing.T) {
	type args struct {
		data     []int32
		lessFunc func(a, b int32) bool
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			"normal",
			args{
				data:     []int32{1, 2, 3},
				lessFunc: gofunctools.Less[int32],
			},
			3,
		},
		{
			"min",
			args{
				data:     []int32{-2147483648, 1, 2, 3},
				lessFunc: gofunctools.Greater[int32],
			},
			-2147483648,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gofunctools.Max(tt.args.lessFunc, tt.args.data...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindMax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapWithLabel(t *testing.T) {
	type args struct {
		data       []string
		leftLabel  string
		rightLabel string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"normal",
			args{
				[]string{"1", "2", "3"},
				"<",
				">",
			},
			"<1><2><3>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gofunctools.WrapWithLabel(tt.args.data, tt.args.leftLabel, tt.args.rightLabel); got != tt.want {
				t.Errorf("WrapWithLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMin(t *testing.T) {
	type args struct {
		data     []int32
		lessFunc func(a, b int32) bool
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			"normal",
			args{
				data:     []int32{1, 2, 3},
				lessFunc: gofunctools.Less[int32],
			},
			1,
		},
		{
			"max",
			args{
				data:     []int32{2147483647, 1, 2, 3},
				lessFunc: gofunctools.Greater[int32],
			},
			2147483647,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gofunctools.Min(tt.args.lessFunc, tt.args.data...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindMin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFormReq(t *testing.T) {
	type Req struct {
		Required string `validate:"required"`
		Optional string `form:",omitempty" validate:"omitempty,min=1"`
	}

	type args struct {
		value     url.Values
		parsedReq *Req
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"normal",
			args{
				value: url.Values{
					"Required": []string{"1"},
					"Optional": []string{"1"},
				},
				parsedReq: &Req{},
			},
			false,
		},
		{
			"miss optional",
			args{
				value: url.Values{
					"Required": []string{"1"},
				},
				parsedReq: &Req{},
			},
			false,
		},
		{
			"miss required",
			args{
				value:     url.Values{},
				parsedReq: &Req{},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := gofunctools.ParseFormReq(tt.args.value, tt.args.parsedReq); (err != nil) != tt.wantErr {
				t.Errorf("ParseFormReq() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
