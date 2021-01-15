package core

import (
	"testing"
)

// func Test_install(t *testing.T) {
// 	type args struct {
// 		queue   []Component
// 		visited []Component
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []Component
// 		wantErr bool
// 	}{
// 		{
// 			"nil",
// 			args{
// 				queue:   nil,
// 				visited: nil,
// 			},
// 			nil,
// 			false,
// 		},

// 		{
// 			"empty",
// 			args{
// 				queue:   []Component{},
// 				visited: []Component{},
// 			},
// 			[]Component{},
// 			false,
// 		},

// 		{
// 			"basic",
// 			args{
// 				queue: []Component{
// 					{
// 						Name:          "git/component-latest",
// 						ComponentType: "component",
// 						Method:        "git",
// 						Source:        "https://github.com/microsoft/fabrikate-definitions",
// 						Path:          "definitions/fabrikate-jaeger",
// 					},
// 					{
// 						Name:          "helm/helm-latest",
// 						ComponentType: "helm",
// 						Method:        "helm",
// 						Source:        "https://grafana.github.io/helm-charts",
// 						Path:          "grafana",
// 					},
// 					{
// 						Name:          "helm/helm-6.1.3",
// 						ComponentType: "helm",
// 						Method:        "helm",
// 						Source:        "https://grafana.github.io/helm-charts",
// 						Version:       "6.1.3",
// 						Path:          "grafana",
// 					},
// 				},
// 				visited: []Component{},
// 			},
// 			[]Component{},
// 			false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := install(tt.args.queue, tt.args.visited)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("install() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("install() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_install(t *testing.T) {
	type args struct {
		queue   []Component
		visited []Component
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"nil",
			args{
				queue:   nil,
				visited: nil,
			},
			false,
		},

		{
			"empty",
			args{
				queue:   []Component{},
				visited: []Component{},
			},
			false,
		},

		{
			"basic",
			args{
				queue: []Component{
					// {
					// 	Name:          "git/component-latest",
					// 	ComponentType: "component",
					// 	Method:        "git",
					// 	Source:        "https://github.com/microsoft/fabrikate-definitions",
					// 	Path:          "definitions/fabrikate-jaeger",
					// },
					{
						Name:          "helm/helm-latest",
						ComponentType: "helm",
						Method:        "helm",
						Source:        "https://grafana.github.io/helm-charts",
						Path:          "grafana",
					},
					{
						Name:          "helm/helm-6.1.3",
						ComponentType: "helm",
						Method:        "helm",
						Source:        "https://grafana.github.io/helm-charts",
						Version:       "6.1.3",
						Path:          "grafana",
					},
				},
				visited: []Component{},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := install(tt.args.queue, tt.args.visited); (err != nil) != tt.wantErr {
				t.Errorf("install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
