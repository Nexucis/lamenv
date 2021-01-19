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
				Title       string
				Field       int
				MyMap       map[string]string `mapstructure:"my_map"`
				Slice       []string
				Array       [1]int
				MyInterface interface{}
			}{},
			result: &ring{
				kind:  root,
				value: "",
				children: []*ring{
					{
						kind:     leaf,
						value:    "Title",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "Field",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "my_map",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "Slice_0",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "Array_0",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "MyInterface",
						children: nil,
					},
				},
			},
		},
		{
			title: "array of struct",
			config: []struct {
				Title       string
				Field       int
				MyMap       map[string]string `mapstructure:"my_map"`
				Slice       []string
				Array       [1]int
				MyInterface interface{} `yaml:"my_interface"`
			}{},
			result: &ring{
				kind:  root,
				value: "_0",
				children: []*ring{
					{
						kind:     leaf,
						value:    "Title",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "Field",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "my_map",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "Slice_0",
						children: nil,
					},
					{
						kind:     leaf,
						value:    "Array_0",
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
				MyInterface interface{} `json:"-" yaml:"-"`
				Title       string      `yaml:",inline"`
				OtherTitle  string      `mapstructure:",inline"`
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
			title: "ignoring unexported field",
			config: struct {
				unexportedFieldTitle string
				ExportedField        uint64
			}{},
			result: &ring{
				kind:  root,
				value: "",
				children: []*ring{
					{
						kind:     leaf,
						value:    "ExportedField",
						children: nil,
					},
				},
			},
		},
		{
			title: "complexe struct",
			config: struct {
				A struct {
					B struct {
						C struct {
							Slice []struct {
								Array [1]struct {
									Leaf map[string]struct{}
								}
							}
						}
					}
				}
				A2 struct {
					Title string
				} `mapstructure:",squash"`
				A3 []struct {
					InnerNode int `yaml:"inner_node"`
				} `mapstructure:",squash"`
			}{},
			result: &ring{
				kind:  root,
				value: "",
				children: []*ring{
					{
						kind:  node,
						value: "A",
						children: []*ring{
							{
								kind:  node,
								value: "B",
								children: []*ring{
									{
										kind:  node,
										value: "C",
										children: []*ring{
											{
												kind:  node,
												value: "Slice_0",
												children: []*ring{
													{
														kind:  node,
														value: "Array_0",
														children: []*ring{
															{
																kind:     leaf,
																value:    "Leaf",
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
								value:    "Title",
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
