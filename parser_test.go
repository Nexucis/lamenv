package lamenv

import (
	"reflect"
	"strings"
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
				kind:     deadLeaf,
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
				value:    "0",
				children: nil,
			},
		},
		{
			title:  "leaf 4",
			config: [1]string{"test"},
			result: &ring{
				kind:     leaf,
				value:    "0",
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
						kind:     deadLeaf,
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
						kind:     deadLeaf,
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
				value: "0",
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
						kind:     deadLeaf,
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
						kind:     deadLeaf,
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
																kind:     deadLeaf,
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
						kind:  nodeSquashed,
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
						kind:  nodeSquashed,
						value: "0",
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

func TestPathPossibility(t *testing.T) {
	testSuites := []struct {
		title  string
		r      *ring
		part   string
		result uint64
	}{
		{
			title: "simple ring",
			r: &ring{
				kind:     leaf,
				value:    "a",
				children: nil,
			},
			part:   "a",
			result: 1,
		},
		{
			title: "simple ring and not matched path",
			r: &ring{
				kind:     leaf,
				value:    "b",
				children: nil,
			},
			part:   "a",
			result: 0,
		},
		{
			title: "simple ring with a value to be aggregated",
			r: &ring{
				kind:     leaf,
				value:    "a_b",
				children: nil,
			},
			part:   "a_b",
			result: 1,
		},
		{
			title: "node squashed",
			r: &ring{
				kind:  nodeSquashed,
				value: "",
				children: []*ring{
					{
						kind:     leaf,
						value:    "a",
						children: nil,
					},
					{
						kind:  nodeSquashed,
						value: "",
						children: []*ring{
							{
								kind:     leaf,
								value:    "a",
								children: nil,
							},
						},
					},
				},
			},
			part:   "a",
			result: 2,
		},
		{
			title: "node squashed",
			r: &ring{
				kind:  nodeSquashed,
				value: "",
				children: []*ring{
					{
						kind:     leaf,
						value:    "a",
						children: nil,
					},
					{
						kind:  nodeSquashed,
						value: "",
						children: []*ring{
							{
								kind:     leaf,
								value:    "a",
								children: nil,
							},
						},
					},
				},
			},
			part:   "a",
			result: 2,
		},
		{
			title: "node squashed 2",
			r: &ring{
				kind:  node,
				value: "a",
				children: []*ring{
					{
						kind:     leaf,
						value:    "a",
						children: nil,
					},
					{
						kind:  nodeSquashed,
						value: "0",
						children: []*ring{
							{
								kind:     leaf,
								value:    "b",
								children: nil,
							},
						},
					},
				},
			},
			part:   "a_0_b",
			result: 1,
		},
		{
			title: "complex tree",
			r: &ring{
				kind:  node,
				value: "a_b",
				children: []*ring{
					{
						kind:  node,
						value: "b_c",
						children: []*ring{
							{
								kind:  nodeSquashed,
								value: "",
								children: []*ring{
									{
										kind:     leaf,
										value:    "c",
										children: nil,
									},
								},
							},
						},
					},
					{
						kind:  nodeSquashed,
						value: "0",
						children: []*ring{
							{
								kind:  node,
								value: "d",
								children: []*ring{
									{
										kind:     leaf,
										value:    "",
										children: nil,
									},
								},
							},
						},
					},
				},
			},
			part:   "a_b_b_c_c",
			result: 1,
		},
	}

	for _, test := range testSuites {
		t.Run(test.title, func(t *testing.T) {
			var r uint64 = 0
			pathPossibility(strings.Split(test.part, "_"), 0, test.r, &r)
			assert.Equal(t, test.result, r)
		})
	}
}

func TestPossiblePrefix(t *testing.T) {
	testSuites := []struct {
		title  string
		part   string
		value  string
		pos    int
		result []possiblePrefix
	}{
		{
			title: "simple match",
			part:  "a",
			value: "a",
			pos:   0,
			result: []possiblePrefix{
				{
					value:    "",
					startPos: 0,
					endPos:   0,
				},
			},
		},
		{
			title:  "no simple matching",
			part:   "b",
			value:  "a",
			pos:    0,
			result: nil,
		},
		{
			title: "recursive matching",
			part:  "a_b_c_b_c_a",
			value: "b_c",
			pos:   0,
			result: []possiblePrefix{
				{
					value:    "a",
					startPos: 1,
					endPos:   2,
				},
				{
					value:    "a_b_c",
					startPos: 3,
					endPos:   4,
				},
			},
		},
	}
	for _, test := range testSuites {
		t.Run(test.title, func(t *testing.T) {
			assert.Equal(t, test.result, findPrefixes(strings.Split(test.part, "_"), test.pos, test.value))
		})
	}
}

func TestGuessPrefix(t *testing.T) {
	testSuites := []struct {
		title  string
		r      *ring
		part   string
		result string
	}{
		{
			title: "leaf",
			r: &ring{
				kind:     leaf,
				value:    "",
				children: nil,
			},
			part:   "a",
			result: "a",
		},
		{
			title: "leaf 2",
			r: &ring{
				kind:     leaf,
				value:    "0",
				children: nil,
			},
			part:   "a_0",
			result: "a",
		},
		{
			title: "leaf 3",
			r: &ring{
				kind:     leaf,
				value:    "0",
				children: nil,
			},
			part:   "a_b_c_d_0",
			result: "a_b_c_d",
		},
		{
			title: "leaf 4",
			r: &ring{
				kind:     leaf,
				value:    "0",
				children: nil,
			},
			part:   "a_b_0_0_0",
			result: "a_b_0_0",
		},
	}
	for _, test := range testSuites {
		t.Run(test.title, func(t *testing.T) {
			prefix, err := guessPrefix(strings.Split(test.part, "_"), test.r)
			assert.NoError(t, err)
			assert.Equal(t, test.result, prefix)
		})
	}
}
