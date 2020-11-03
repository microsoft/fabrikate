package cmd

import (
	"testing"

	"github.com/microsoft/fabrikate/core"
)

func TestGenerate(t *testing.T) {
	type args struct {
		startPath    string
		environments []string
		validate     bool
	}
	tests := []struct {
		name        string
		args        args
		wantLengths map[string]int
		wantErr     bool
	}{
		{
			"json",
			args{
				"../testdata/generate",
				[]string{"prod-east", "prod"},
				false,
			},
			map[string]int{
				"microservices-workload": 0,
				"infra":                  0,
				"fabrikate-jaeger":       409,
				"jaeger":                 26877,
			},
			false,
		},

		{
			"yaml",
			args{
				"../testdata/generate-yaml",
				[]string{"prod"},
				false,
			},
			map[string]int{
				"prometheus-grafana": 125,
				"grafana":            8552,
				"prometheus":         29370,
			},
			false,
		},

		{
			"remote static",
			args{
				"../testdata/generate-remote-static",
				[]string{"common"},
				false,
			},
			map[string]int{
				"keyvault-flexvolume": 5,
				"keyvault-sub":        1372,
			},
			false,
		},

		{
			"hooks",
			args{
				"../testdata/generate-hooks",
				[]string{"prod"},
				false,
			},
			map[string]int{
				"generate-hooks": 103,
			},
			false,
		},

		{
			"disabled subcomponent",
			args{
				"../testdata/generate-disabled",
				[]string{"disabled"},
				false,
			},
			map[string]int{
				"disabled-stack": 0,
			},
			false,
		},
	}

	checkComponentLengthsAgainstExpected := func(t *testing.T, components []core.Component, expectedLengths map[string]int) {
		for _, component := range components {
			if expectedLength, ok := expectedLengths[component.Name]; ok {
				actualLength := len(component.Manifest)
				if actualLength != expectedLength {
					t.Errorf("Generate() manifest length %v, want %v", actualLength, expectedLength)
				}
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotComponents, err := Generate(tt.args.startPath, tt.args.environments, tt.args.validate)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			checkComponentLengthsAgainstExpected(t, gotComponents, tt.wantLengths)
		})
	}
}
