package cmd

import (
	"testing"
)

func TestInstall(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"json",
			args{"../../testdata/install"},
			false,
		},

		{
			"yaml",
			args{"../../testdata/install-yaml"},
			false,
		},

		{
			"hooks",
			args{"../../testdata/install-hooks"},
			false,
		},

		{
			"private git",
			args{"../../testdata/install-private"},
			true,
		},

		{
			"helm",
			args{"../../testdata/install-helm"},
			false,
		},

		{
			"repo-alias",
			args{"../../testdata/repo-alias"},
			false,
		},
	}

	for _, tt := range tests {
		defer func() {
			// 	_ = util.UninstallComponents(tt.args.path)
		}()

		t.Run(tt.name, func(t *testing.T) {
			if err := Install(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
