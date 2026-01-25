package filesystem

import (
	"os"
	"testing"
)

func TestCreateDirIfNotExists(t *testing.T) {

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create new directory",
			path:    "testdata",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.Remove(tt.path)
			if err := CreateDirIfNotExists(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("CreateDirIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if _, err := os.Stat(tt.path); os.IsNotExist(err) {
				t.Errorf("Directory %s was not created", tt.path)
			}
		})
	}
}

func Test_dirExists(t *testing.T) {
	type args struct {
		path       string
		needCreate bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "directory exists",
			args: args{path: "testdata_exist", needCreate: true},
			want: true,
		},
		{
			name: "directory does not exist",
			args: args{path: "testdata_not_exist", needCreate: false},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.needCreate {
				err := os.MkdirAll(tt.args.path, os.ModePerm)
				if err != nil {
					t.Errorf("MkdirAll() error = %v", err)
				}
				defer os.RemoveAll(tt.args.path)
			}
			if got := dirExists(tt.args.path); got != tt.want {
				t.Errorf("dirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
