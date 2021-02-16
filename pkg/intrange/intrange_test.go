package intrange

import "testing"

func TestIntRange_Get(t *testing.T) {
	type fields struct {
		min int
		max int
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "normal range true", fields: fields{0, 10}, args: args{5}, want: true},
		{name: "normal start range true", fields: fields{0, 10}, args: args{0}, want: true},
		{name: "normal end range true", fields: fields{0, 10}, args: args{10}, want: true},
		{name: "small range true", fields: fields{1, 1}, args: args{1}, want: true},
		{name: "normal start range false", fields: fields{1, 2}, args: args{0}, want: false},
		{name: "normal end range false", fields: fields{1, 2}, args: args{3}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := IntRange{
				min: tt.fields.min,
				max: tt.fields.max,
			}
			if got := r.Get(tt.args.n); got != tt.want {
				t.Errorf("IntRange.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntRanges_Get(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		rs   IntRanges
		args args
		want bool
	}{
		{name: "normal range true", rs: IntRanges{{0, 10}}, args: args{5}, want: true},
		{name: "normal ranges inbetween true", rs: IntRanges{{0, 4}, {5, 10}}, args: args{5}, want: true},
		{name: "normal ranges inbetween false", rs: IntRanges{{0, 4}, {6, 10}}, args: args{5}, want: false},
		{name: "normal start range true", rs: IntRanges{{0, 10}}, args: args{0}, want: true},
		{name: "normal end range true", rs: IntRanges{{0, 10}}, args: args{10}, want: true},
		{name: "small range true", rs: IntRanges{{1, 1}, {3, 3}}, args: args{1}, want: true},
		{name: "normal start range false", rs: IntRanges{{1, 2}}, args: args{0}, want: false},
		{name: "normal end range false", rs: IntRanges{{1, 2}}, args: args{3}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rs.Get(tt.args.n); got != tt.want {
				t.Errorf("IntRanges.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
