package crypto

import (
	"github.com/tj/assert"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "#1",
			args: args{"1234567812345678"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.True(t, reflect.TypeOf(got).Kind() == reflect.String)
			assert.True(t, len(got) > 10)

			got2, err := Decode(got)

			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, got2, tt.args.str)

		})
	}
}

func Test_generateRandom(t *testing.T) {
	type args struct {
		size int
	}
	var tests = []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "#1",
			args: args{10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateRandom(tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateRandom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.True(t, len(got) == 10)

		})
	}
}

func Test_keyInit(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Test key init",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := keyInit(); (err != nil) != tt.wantErr {
				t.Errorf("keyInit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
