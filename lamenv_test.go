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
