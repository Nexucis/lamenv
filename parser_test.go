package lamenv

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testSuites := []struct {
		title  string
		config interface{}
		result *ring
	}{
		{
			title:  "leaf",
			config: map[string]string{},
			result: &ring{
				kind:     leaf,
				value:    "",
				children: nil,
			},
		},
		{
			title:  "leaf2",
			config: 3,
			result: &ring{
				kind:     leaf,
				value:    "",
				children: nil,
			},
		},
		{
			title:  "leaf 3",
			config: []string{"test"},
			result: &ring{
				kind:     leaf,
				value:    "_0",
				children: nil,
			},
		},
		{
			title:  "leaf 4",
			config: [1]string{"test"},
			result: &ring{
				kind:     leaf,
				value:    "_0",
				children: nil,
			},
		},
		{
			title:  "leaf 5",
			config: "test",
			result: &ring{
				kind:     leaf,
				value:    "",
				children: nil,
			},
		},
		{
			title: "struct with a single level",
			config: struct {
				title       string
				field       int
				myMap       map[string]string `mapstructure:"my_map"`
				slice       []string
				array       [1]int
				myInterface interface{}
			}{},
			result: &ring{
				kind:  root,
				value: "",
				children: []*ring{
					{
						kind:     leaf,
						value:    "title",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "field",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "my_map",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "slice_0",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "array_0",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "myInterface",
						children: nil,
					},
				},
			},
		},
		{
			title: "array of struct",
			config: []struct {
				title       string
				field       int
				myMap       map[string]string `mapstructure:"my_map"`
				slice       []string
				array       [1]int
				myInterface interface{} `yaml:"my_interface"`
			}{},
			result: &ring{
				kind:  root,
				value: "_0",
				children: []*ring{
					{
						kind:     leaf,
						value:    "title",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "field",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "my_map",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "slice_0",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "array_0",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "my_interface",
						children: nil,
					},
				},
			},
		},
		{
			title: "ignoring and squashing item in struct",
			config: struct {
				myInterface interface{} `json:"-" yaml:"-"`
				title       string      `yaml:",inline"`
				otherTitle  string      `mapstructure:",inline"`
			}{},
			result: &ring{
				kind:  root,
				value: "",
				children: []*ring{
					{
						kind:     leaf,
						value:    "",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "",
						children: nil,
					},
				},
			},
		},
		{
			title: "complexe struct",
			config: struct {
				a struct {
					b struct {
						c struct {
							slice []struct {
								array [1]struct {
									leaf map[string]struct{}
								}
							}
						}
					}
				}
				a2 struct {
					title string
				} `mapstructure:",squash"`
				a3 []struct {
					innerNode int `yaml:"inner_node"`
				} `mapstructure:",squash"`
			}{},
			result: &ring{
				kind:  root,
				value: "",
				children: []*ring{
					{
						kind:  node,
						value: "a",
						children: []*ring{
							{
								kind:  node,
								value: "b",
								children: []*ring{
									{
										kind:  node,
										value: "c",
										children: []*ring{
											{
												kind:  node,
												value: "slice_0",
												children: []*ring{
													{
														kind:  node,
														value: "array_0",
														children: []*ring{
															{
																kind:     leaf,
																value:    "leaf",
																children: nil,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						kind:  node,
						value: "",
						children: []*ring{
							{
								kind:     leaf,
								value:    "title",
								children: nil,
							},
						},
					},
					{
						kind:  node,
						value: "_0",
						children: []*ring{
							{
								kind:     leaf,
								value:    "inner_node",
								children: nil,
							},
						},
					},
				},
			},
		},
	}
	for _, test := range testSuites {
		t.Run(test.title, func(t *testing.T) {
			v := reflect.TypeOf(test.config)
			assert.Equal(t, test.result, newRing(v, defaultTagSupported))
		})
	}
}
