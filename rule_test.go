package exceltable

import (
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestRegisterRule(t *testing.T) {
	RegisterRule(-1000, "tmp", &excelize.Style{})
}

func TestRegisterPredicate(t *testing.T) {
	RegisterPredicate("tmp", func() bool { return true })
}

func TestCountByRule(t *testing.T) {
	type args struct {
		obj *person
		tag string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Positive",
			args: args{
				obj: persons[0],
				tag: "warn",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Posotive: Custom rule",
			args: args{
				obj: persons[0],
				tag: "newface",
			},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountByRule(tt.args.obj, tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CountByRule() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("CountByRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_callPredicate(t *testing.T) {
	type args struct {
		pred reflect.Value
		arg  reflect.Value
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Positive: nulary predicate",
			args: args{
				pred: reflect.ValueOf(func() bool { return true }),
				arg:  reflect.ValueOf(nil),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Positive: nulary predicate",
			args: args{
				pred: reflect.ValueOf(func() bool { return false }),
				arg:  reflect.ValueOf(nil),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Positive: unary predicate",
			args: args{
				pred: reflect.ValueOf(func(s string) bool { return s == "something arg" }),
				arg:  reflect.ValueOf("something arg"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Positive: unary predicate",
			args: args{
				pred: reflect.ValueOf(func(s string) bool { return s == "something arg" }),
				arg:  reflect.ValueOf("another arg"),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Negative: not predicate",
			args: args{
				pred: reflect.ValueOf(func() int { return 1 }),
				arg:  reflect.ValueOf(nil),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Negative: two aruments",
			args: args{
				pred: reflect.ValueOf(func(_, _ string) bool { return true }),
				arg:  reflect.ValueOf(nil),
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := callPredicate(tt.args.pred, tt.args.arg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("callPredicate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("callPredicate() = %v, want %v", got, tt.want)
			}
		})
	}
}
