package lamenv

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshallEnv(t *testing.T) {
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
			title: "slice with native type 2",
			config: &struct {
				Slice []int
			}{},
			env: map[string]string{
				"SLICE": "3, 2",
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
