package lamenv

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	duration3h    = 3 * time.Hour
	durationEmpty = time.Duration(0)
)

// This type is a dummy alias of string which proves that encoding.TextUnmarshaler and encoding.TextMarshaler can be
// implemented for custom encoding/decoding.
// It replaces "foo" with "bar" when unmarshalling and "bar" with "foo" when marshalling.
type dummyString string

func (s *dummyString) UnmarshalText(text []byte) error {
	*s = dummyString(strings.Replace(string(text), "foo", "bar", -1))
	return nil
}

func (s *dummyString) MarshalText() ([]byte, error) {
	return []byte(strings.Replace(string(*s), "bar", "foo", -1)), nil
}

func TestUnmarshalWithOverriding(t *testing.T) {
	type InlineStruct struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	type complexStruct struct {
		InlineStruct `json:",inline"`
		C            string `json:"c"`
	}
	type finalStruct struct {
		Slice []*complexStruct `json:"slice"`
		D     string           `json:"d"`
		F     string           `json:"f"`
	}

	testSuite := []struct {
		title  string
		jason  string
		env    map[string]string
		result finalStruct
	}{
		{
			title: "",
			jason: `
{
  "d": "not empty",
  "slice": [
    {
      "a": "not empty"
    }
  ]
}
`,
			env: map[string]string{
				"SLICE_0_B": "from env",
				"SLICE_1_A": "from env",
			},
			result: finalStruct{
				Slice: []*complexStruct{
					{
						InlineStruct: InlineStruct{
							A: "not empty",
							B: "from env",
						},
						C: "",
					},
					{
						InlineStruct: InlineStruct{
							A: "from env",
						},
					},
				},
				D: "not empty",
				F: "",
			},
		},
	}

	for _, test := range testSuite {
		t.Run(test.title, func(t *testing.T) {
			for k, v := range test.env {
				_ = os.Setenv(k, v)
			}
			c := &finalStruct{}
			assert.NoError(t, json.Unmarshal([]byte(test.jason), c))
			assert.NoError(t, Unmarshal(c, []string{}))
			assert.Equal(t, test.result, *c)
			for k := range test.env {
				_ = os.Unsetenv(k)
			}
		})
	}

}

func TestUnmarshal(t *testing.T) {
	testSuites := []struct {
		title  string
		parts  []string
		config interface{}
		env    map[string]string
		result interface{}
	}{
		{
			title: "simple config with native type",
			config: &struct {
				A     int
				A1    int8
				A2    int16
				A3    int32
				B     uint
				B1    uint
				B2    uint
				B3    uint
				C     bool
				Title string
			}{},
			env: map[string]string{
				"A":     "78945613",
				"A1":    "-1",
				"A2":    "-123",
				"A3":    "456789",
				"B":     "78945613",
				"B1":    "1",
				"B2":    "123",
				"B3":    "456789",
				"C":     "true",
				"TITLE": "my title",
			},
			result: &struct {
				A     int
				A1    int8
				A2    int16
				A3    int32
				B     uint
				B1    uint
				B2    uint
				B3    uint
				C     bool
				Title string
			}{
				A:     78945613,
				A1:    -1,
				A2:    -123,
				A3:    456789,
				B:     78945613,
				B1:    1,
				B2:    123,
				B3:    456789,
				C:     true,
				Title: "my title",
			},
		},
		{
			title: "same simple config with native type with prefix",
			config: &struct {
				A     int
				A1    int8
				A2    int16
				A3    int32
				B     uint
				B1    uint
				B2    uint
				B3    uint
				C     bool
				Title string
			}{},
			parts: []string{"PREFIX"},
			env: map[string]string{
				"PREFIX_A":     "78945613",
				"PREFIX_A1":    "-1",
				"PREFIX_A2":    "-123",
				"PREFIX_A3":    "456789",
				"PREFIX_B":     "78945613",
				"PREFIX_B1":    "1",
				"PREFIX_B2":    "123",
				"PREFIX_B3":    "456789",
				"PREFIX_C":     "true",
				"PREFIX_TITLE": "my title",
			},
			result: &struct {
				A     int
				A1    int8
				A2    int16
				A3    int32
				B     uint
				B1    uint
				B2    uint
				B3    uint
				C     bool
				Title string
			}{
				A:     78945613,
				A1:    -1,
				A2:    -123,
				A3:    456789,
				B:     78945613,
				B1:    1,
				B2:    123,
				B3:    456789,
				C:     true,
				Title: "my title",
			},
		},
		{
			title: "same simple config with multiple tag support",
			config: &struct {
				A     int   `yaml:"a"`
				A1    int8  `json:"a_1"`
				A2    int16 `mapstructure:"a_2"`
				A3    int32 `yaml:"a_3"`
				B     uint
				B1    uint
				B2    uint
				B3    uint
				C     bool
				Title string
			}{},
			parts: []string{"PREFIX"},
			env: map[string]string{
				"PREFIX_A":     "78945613",
				"PREFIX_A_1":   "-1",
				"PREFIX_A_2":   "-123",
				"PREFIX_A_3":   "456789",
				"PREFIX_B":     "78945613",
				"PREFIX_B1":    "1",
				"PREFIX_B2":    "123",
				"PREFIX_B3":    "456789",
				"PREFIX_C":     "true",
				"PREFIX_TITLE": "my title",
			},
			result: &struct {
				A     int   `yaml:"a"`
				A1    int8  `json:"a_1"`
				A2    int16 `mapstructure:"a_2"`
				A3    int32 `yaml:"a_3"`
				B     uint
				B1    uint
				B2    uint
				B3    uint
				C     bool
				Title string
			}{
				A:     78945613,
				A1:    -1,
				A2:    -123,
				A3:    456789,
				B:     78945613,
				B1:    1,
				B2:    123,
				B3:    456789,
				C:     true,
				Title: "my title",
			},
		},
		{
			title: "inner struct",
			config: &struct {
				Aptr *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"a_ptr"`
				A struct {
					A struct {
						A struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}
					}
				}
			}{},
			parts: []string{"PREFIX"},
			env: map[string]string{
				"PREFIX_A_PTR_INNER_NODE":       "1",
				"PREFIX_A_A_A_SUPER_INNER_NODE": "2",
			},
			result: &struct {
				Aptr *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"a_ptr"`
				A struct {
					A struct {
						A struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}
					}
				}
			}{
				Aptr: &struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					InnerNode: 1,
				},
				A: struct {
					A struct {
						A struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}
					}
				}{
					A: struct {
						A struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}
					}{
						A: struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}{
							SuperInnerNode: 2,
						},
					},
				},
			},
		},
		{
			title: "omitempty management for pointer",
			config: &struct {
				Ptr1 *struct {
					InnerNode int
				} `mapstructure:"ptr1,omitempty"` // the outcome is that the pointer shouldn't be initialized if the env doesn't exist
				Ptr2 *struct {
					InnerNode int
				} `mapstructure:"ptr2"` // here the pointer should be created even if the env doesn't exist
			}{},
			result: &struct {
				Ptr1 *struct {
					InnerNode int
				} `mapstructure:"ptr1,omitempty"`
				Ptr2 *struct {
					InnerNode int
				} `mapstructure:"ptr2"`
			}{
				Ptr2: &struct {
					InnerNode int
				}{},
			},
		},
		{
			title: "omitempty management for pointer 2",
			config: &struct {
				Ptr1 *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"ptr_1,omitempty"`
			}{},
			env: map[string]string{
				"PTR_1_INNER_NODE": "3",
			},
			result: &struct {
				Ptr1 *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"ptr_1,omitempty"`
			}{
				Ptr1: &struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					InnerNode: 3,
				},
			},
		},
		{
			title: "omitempty management for pointer 3",
			config: &struct {
				Ptr1 *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"ptr_1,  omitempty"` // so much space here, but it shouldn't cause an issue
			}{},
			env: map[string]string{
				"PTR_1_INNER_NODE": "3",
			},
			result: &struct {
				Ptr1 *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"ptr_1,  omitempty"`
			}{
				Ptr1: &struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					InnerNode: 3,
				},
			},
		},
		{
			title: "slice with native type",
			config: &struct {
				Slice []int
			}{},
			env: map[string]string{
				"SLICE_0": "3",
				"SLICE_1": "2",
			},
			result: &struct {
				Slice []int
			}{
				Slice: []int{3, 2},
			},
		},
		{
			title: "slice of struct",
			config: &struct {
				Slice []struct {
					InnerNode int `mapstructure:"inner_node"`
				}
			}{},
			env: map[string]string{
				"SLICE_0_INNER_NODE": "5",
				"SLICE_1_INNER_NODE": "1",
			},
			result: &struct {
				Slice []struct {
					InnerNode int `mapstructure:"inner_node"`
				}
			}{
				Slice: []struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					{
						InnerNode: 5,
					},
					{
						InnerNode: 1,
					},
				},
			},
		},
		{
			title: "slice of pointer to struct",
			config: &struct {
				Slice []*struct {
					InnerNode int `mapstructure:"inner_node"`
				}
			}{},
			env: map[string]string{
				"SLICE_0_INNER_NODE": "5",
				"SLICE_1_INNER_NODE": "1",
			},
			result: &struct {
				Slice []*struct {
					InnerNode int `mapstructure:"inner_node"`
				}
			}{
				Slice: []*struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					{
						InnerNode: 5,
					},
					{
						InnerNode: 1,
					},
				},
			},
		},
		{
			title: "map of native type",
			config: &struct {
				Map  map[string]int
				Map2 map[string]float64
			}{},
			env: map[string]string{
				"MAP_LOL":        "5",
				"MAP2_SUPER_FUN": "1",
			},
			result: &struct {
				Map  map[string]int
				Map2 map[string]float64
			}{
				Map: map[string]int{
					"lol": 5,
				},
				Map2: map[string]float64{
					"super_fun": 1,
				},
			},
		},
		{
			title: "map of complex type",
			config: &struct {
				Map map[string]struct {
					InnerNode int `mapstructure:"inner_node"`
				}
				Map2 map[string][]int
			}{},
			env: map[string]string{
				"MAP_LOL_INNER_NODE": "5",
				"MAP_PGM_INNER_NODE": "6",
				"MAP2_SUPER_FUN_0":   "1",
				"MAP2_SUPER_FUN_1":   "2",
				"MAP2_FUN_0":         "4",
				"MAP2_FUN_1":         "5",
			},
			result: &struct {
				Map map[string]struct {
					InnerNode int `mapstructure:"inner_node"`
				}
				Map2 map[string][]int
			}{
				Map: map[string]struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					"lol": {InnerNode: 5},
					"pgm": {InnerNode: 6},
				},
				Map2: map[string][]int{
					"super_fun": {1, 2},
					"fun":       {4, 5},
				},
			},
		},
		{
			title: "map of complex type 2",
			config: &struct {
				Map map[string]struct {
					My struct {
						Key       string
						InnerNode []struct {
							Map map[string]string
						} `mapstructure:"inner_node"`
					}
				}
			}{},
			env: map[string]string{
				"MAP_MY_KEY_MY_KEY":                        "lol",
				"MAP_MY_MY_MY_KEY":                         "gg",
				"MAP_MY_MY_MY_INNER_NODE_0_MAP_INNER_NODE": "5",
				"MAP_MY_MY_MY_MY_KEY":                      "gg",
			},
			result: &struct {
				Map map[string]struct {
					My struct {
						Key       string
						InnerNode []struct {
							Map map[string]string
						} `mapstructure:"inner_node"`
					}
				}
			}{
				Map: map[string]struct {
					My struct {
						Key       string
						InnerNode []struct{ Map map[string]string } `mapstructure:"inner_node"`
					}
				}{
					"my_key": {
						My: struct {
							Key       string
							InnerNode []struct{ Map map[string]string } `mapstructure:"inner_node"`
						}{
							Key: "lol",
						},
					},
					"my_my": {
						My: struct {
							Key       string
							InnerNode []struct{ Map map[string]string } `mapstructure:"inner_node"`
						}{
							Key: "gg",
							InnerNode: []struct{ Map map[string]string }{
								{
									Map: map[string]string{
										"inner_node": "5",
									},
								},
							},
						},
					},
					"my_my_my": {
						My: struct {
							Key       string
							InnerNode []struct{ Map map[string]string } `mapstructure:"inner_node"`
						}{
							Key: "gg",
						},
					},
				},
			},
		},
		{
			title: "squash",
			config: &struct {
				S struct {
					A int
					B uint
				} `mapstructure:",inline"`
				S2 struct {
					C float64 `mapstructure:"c,omitempty"`
					D string
				} `mapstructure:",squash"`
			}{},
			env: map[string]string{
				"A": "0",
				"B": "0",
				"D": "awesome string",
			},
			result: &struct {
				S struct {
					A int
					B uint
				} `mapstructure:",inline"`
				S2 struct {
					C float64 `mapstructure:"c,omitempty"`
					D string
				} `mapstructure:",squash"`
			}{
				S: struct {
					A int
					B uint
				}{
					A: 0,
					B: 0,
				},
				S2: struct {
					C float64 `mapstructure:"c,omitempty"`
					D string
				}{
					D: "awesome string",
				},
			},
		},
		{
			title: "time.Duration",
			config: &struct {
				A time.Duration
				B *time.Duration
				C time.Duration
				D *time.Duration
				E time.Duration  `mapstructure:"e,omitempty"`
				F *time.Duration `mapstructure:"f,omitempty"`
			}{},
			env: map[string]string{
				"A": "60s",
				"B": "3h",
			},
			result: &struct {
				A time.Duration
				B *time.Duration
				C time.Duration
				D *time.Duration
				E time.Duration  `mapstructure:"e,omitempty"`
				F *time.Duration `mapstructure:"f,omitempty"`
			}{
				A: 1 * time.Minute,
				B: &duration3h,
				C: durationEmpty,
				D: &durationEmpty,
				E: durationEmpty,
				F: &durationEmpty,
			},
		},
		{
			title: "encoding.TextUnmarshaler",
			config: &struct {
				A dummyString
			}{},
			env: map[string]string{
				"A": "foofoofoo",
			},
			result: &struct {
				A dummyString
			}{
				A: "barbarbar",
			},
		},
	}
	for _, test := range testSuites {
		t.Run(test.title, func(t *testing.T) {
			// set env
			for k, v := range test.env {
				_ = os.Setenv(k, v)
			}
			err := Unmarshal(test.config, test.parts)
			assert.Nil(t, err)
			assert.Equal(t, test.result, test.config)
			// unset env
			for k := range test.env {
				_ = os.Unsetenv(k)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	testSuite := []struct {
		title  string
		conf   interface{}
		parts  []string
		result map[string]string
		// resultNotExist is a list of environment variable that shouldn't exist
		// it's typically for testing the omitempty part
		resultNotExist []string
	}{
		{
			title: "native type",
			conf:  5,
			parts: []string{"MY_PREFIX"},
			result: map[string]string{
				"MY_PREFIX": "5",
			},
		},
		{
			title: "simple config",
			conf: &struct {
				A     int
				A1    int8
				A2    int16
				A3    int32
				B     uint
				B1    uint
				B2    uint
				B3    uint
				C     bool
				D     float64
				Title string
			}{
				A:     78945613,
				A1:    -1,
				A2:    -123,
				A3:    456789,
				B:     78945613,
				B1:    1,
				B2:    123,
				B3:    456789,
				C:     true,
				D:     1,
				Title: "my title",
			},
			result: map[string]string{
				"A":     "78945613",
				"A1":    "-1",
				"A2":    "-123",
				"A3":    "456789",
				"B":     "78945613",
				"B1":    "1",
				"B2":    "123",
				"B3":    "456789",
				"C":     "true",
				"D":     "1.000000",
				"TITLE": "my title",
			},
		},
		{
			title: "simple config with multiple tag support",
			conf: &struct {
				A     int
				A1    int8  `json:"a_1"`
				A2    int16 `yaml:"a_2"`
				A3    int32 `mapstructure:"a_3"`
				B     uint
				B1    uint
				B2    uint `json:"b_2" yaml:"b_2" mapstructure:"b_2"`
				B3    uint
				C     bool
				D     float64
				Title string
			}{
				A:     78945613,
				A1:    -1,
				A2:    -123,
				A3:    456789,
				B:     78945613,
				B1:    1,
				B2:    123,
				B3:    456789,
				C:     true,
				D:     1,
				Title: "my title",
			},
			result: map[string]string{
				"A":     "78945613",
				"A_1":   "-1",
				"A_2":   "-123",
				"A_3":   "456789",
				"B":     "78945613",
				"B1":    "1",
				"B_2":   "123",
				"B3":    "456789",
				"C":     "true",
				"D":     "1.000000",
				"TITLE": "my title",
			},
		},
		{
			title: "inner struct",
			parts: []string{"PREFIX"},
			conf: &struct {
				Aptr *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"a_ptr"`
				A struct {
					A struct {
						A struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}
					}
				}
			}{
				Aptr: &struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					InnerNode: 1,
				},
				A: struct {
					A struct {
						A struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}
					}
				}{
					A: struct {
						A struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}
					}{
						A: struct {
							SuperInnerNode int `mapstructure:"super_inner_node"`
						}{
							SuperInnerNode: 2,
						},
					},
				},
			},
			result: map[string]string{
				"PREFIX_A_PTR_INNER_NODE":       "1",
				"PREFIX_A_A_A_SUPER_INNER_NODE": "2",
			},
		},
		{
			title: "slice with native type",
			conf: &struct {
				Slice []int
			}{
				Slice: []int{3, 2},
			},
			result: map[string]string{
				"SLICE_0": "3",
				"SLICE_1": "2",
			},
		},
		{
			title: "slice of struct",
			conf: &struct {
				Slice []struct {
					InnerNode int `mapstructure:"inner_node"`
				}
			}{
				Slice: []struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					{
						InnerNode: 5,
					},
					{
						InnerNode: 1,
					},
				},
			},
			result: map[string]string{
				"SLICE_0_INNER_NODE": "5",
				"SLICE_1_INNER_NODE": "1",
			},
		},
		{
			title: "slice of pointer to struct",
			conf: &struct {
				Slice []*struct {
					InnerNode int `mapstructure:"inner_node"`
				}
			}{
				Slice: []*struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					{
						InnerNode: 5,
					},
					{
						InnerNode: 1,
					},
				},
			},
			result: map[string]string{
				"SLICE_0_INNER_NODE": "5",
				"SLICE_1_INNER_NODE": "1",
			},
		},
		{
			title: "map of native type",
			conf: &struct {
				Map  map[string]int
				Map2 map[string]float64
			}{
				Map: map[string]int{
					"lol": 5,
				},
				Map2: map[string]float64{
					"super_fun": 1,
				},
			},
			result: map[string]string{
				"MAP_LOL":        "5",
				"MAP2_SUPER_FUN": "1.000000",
			},
		},
		{
			title: "map of complex type",
			conf: &struct {
				Map map[string]struct {
					InnerNode int `mapstructure:"inner_node"`
				}
				Map2 map[string][]int
			}{
				Map: map[string]struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					"lol": {InnerNode: 5},
					"pgm": {InnerNode: 6},
				},
				Map2: map[string][]int{
					"super_fun": {1, 2},
					"fun":       {4, 5},
				},
			},
			result: map[string]string{
				"MAP_LOL_INNER_NODE": "5",
				"MAP_PGM_INNER_NODE": "6",
				"MAP2_SUPER_FUN_0":   "1",
				"MAP2_SUPER_FUN_1":   "2",
				"MAP2_FUN_0":         "4",
				"MAP2_FUN_1":         "5",
			},
		},
		{
			title: "map of complex type 2",
			conf: &struct {
				Map map[string]struct {
					My struct {
						Key       string
						InnerNode []struct {
							Map map[string]string
						} `mapstructure:"inner_node"`
					}
				}
			}{
				Map: map[string]struct {
					My struct {
						Key       string
						InnerNode []struct{ Map map[string]string } `mapstructure:"inner_node"`
					}
				}{
					"my_key": {
						My: struct {
							Key       string
							InnerNode []struct{ Map map[string]string } `mapstructure:"inner_node"`
						}{
							Key: "lol",
						},
					},
					"my_my": {
						My: struct {
							Key       string
							InnerNode []struct{ Map map[string]string } `mapstructure:"inner_node"`
						}{
							Key: "gg",
							InnerNode: []struct{ Map map[string]string }{
								{
									Map: map[string]string{
										"inner_node": "5",
									},
								},
							},
						},
					},
					"my_my_my": {
						My: struct {
							Key       string
							InnerNode []struct{ Map map[string]string } `mapstructure:"inner_node"`
						}{
							Key: "gg",
						},
					},
				},
			},
			result: map[string]string{
				"MAP_MY_KEY_MY_KEY":                        "lol",
				"MAP_MY_MY_MY_KEY":                         "gg",
				"MAP_MY_MY_MY_INNER_NODE_0_MAP_INNER_NODE": "5",
				"MAP_MY_MY_MY_MY_KEY":                      "gg",
			},
		},
		{
			title: "omitempty management",
			conf: &struct {
				Slice []string `mapstructure:"slice,omitempty"`
				Ptr1  *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"ptr_1,omitempty"`
				Ptr2 *struct {
					InnerNode int `mapstructure:"inner_node"`
				} `mapstructure:"ptr_2,omitempty"`
				Title string            `yaml:"title,omitempty"`
				Map   map[string]string `json:"map,omitempty"`
				A     uint              `json:"A,omitempty"`
				B     int               `json:"B,omitempty"`
				C     float64           `json:"C,omitempty"`
				D     struct {
					b uint
					A int `json:"A,omitempty"`
				} `json:"D,omitempty"`
				E bool `json:"e,omitempty"`
			}{
				Ptr2: &struct {
					InnerNode int `mapstructure:"inner_node"`
				}{
					InnerNode: 2,
				},
			},
			result: map[string]string{
				"PTR_2_INNER_NODE": "2",
			},
			resultNotExist: []string{
				"SLICE",
				"PTR_1_INNER_NODE",
				"TITLE",
				"MAP",
				"A",
				"B",
				"C",
				"D_A",
				"D_b",
				"E",
			},
		},
		{
			title: "squash",
			conf: &struct {
				S struct {
					A int
					B uint
				} `mapstructure:",inline"`
				S2 struct {
					C float64 `mapstructure:"c,omitempty"`
					D string
				} `mapstructure:",squash"`
			}{
				S: struct {
					A int
					B uint
				}{
					A: 0,
					B: 0,
				},
				S2: struct {
					C float64 `mapstructure:"c,omitempty"`
					D string
				}{
					D: "awesome string",
				},
			},
			result: map[string]string{
				"A": "0",
				"B": "0",
				"D": "awesome string",
			},
			resultNotExist: []string{
				"C",
			},
		},
		{
			title: "time.Duration",
			conf: &struct {
				A time.Duration
				B *time.Duration
				C time.Duration
				D *time.Duration
				E time.Duration  `mapstructure:"e,omitempty"`
				F *time.Duration `mapstructure:"f,omitempty"`
				G *time.Duration `mapstructure:"g,omitempty"`
			}{
				A: 1 * time.Minute,
				B: &duration3h,
				C: durationEmpty,
				D: &durationEmpty,
				E: durationEmpty,
				F: &durationEmpty,
				G: nil,
			},
			result: map[string]string{
				"A": "1m0s",
				"B": "3h0m0s",
				"C": "0s",
				"D": "0s",
				"F": "0s", // a zero value in a pointer is considered as an explicit set so not omitted
			},
			resultNotExist: []string{
				"E",
				"G",
			},
		},
		{
			title: "encoding.TextMarshaler",
			conf: &struct {
				A dummyString
			}{
				A: "barbarbar",
			},
			result: map[string]string{
				"A": "foofoofoo",
			},
		},
	}
	for _, test := range testSuite {
		t.Run(test.title, func(t *testing.T) {
			err := Marshal(test.conf, test.parts)
			assert.NoError(t, err)
			for k, v := range test.result {
				assert.Equal(t, v, os.Getenv(k))
			}
			for _, e := range test.resultNotExist {
				_, ok := os.LookupEnv(e)
				assert.False(t, ok, fmt.Sprintf("variable %s exists and it shouldn't", e))
			}
			for k := range test.result {
				_ = os.Unsetenv(k)
			}
		})
	}
}

func TestLamenv_AddTagSupport(t *testing.T) {
	lam := New()
	lam.AddTagSupport("env")
	assert.Equal(t, append(defaultTagSupported, "env"), lam.tagSupports)
}

func TestLamenv_OverrideTagSupport(t *testing.T) {
	lam := New()
	lam.OverrideTagSupport("env")
	assert.Equal(t, []string{"env"}, lam.tagSupports)
}
