package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_should_parse_valid_config(t *testing.T) {
	testCases := map[string]struct {
		args           []string
		expectedConfig Config
	}{
		"default config": {
			args: []string{},
			expectedConfig: Config{
				addr:              "localhost:8080",
				folder:            "./public",
				baseURL:           "/",
				maxInlineFileSize: 1_048_576,
			},
		},
		"config from cmd": {
			args: []string{
				"--addr",
				":8081",
				"--folder",
				"/tmp/share",
				"--base-url",
				"/lsgo/",
				"--max-inline-file-size",
				"10000000",
			},
			expectedConfig: Config{
				addr:              ":8081",
				folder:            "/tmp/share",
				baseURL:           "/lsgo/",
				maxInlineFileSize: 10000000,
			},
		},
		"adds missing slashes to base url": {
			args: []string{
				"--base-url",
				"lsgo",
			},
			expectedConfig: Config{
				addr:              "localhost:8080",
				folder:            "./public",
				baseURL:           "/lsgo/",
				maxInlineFileSize: 1_048_576,
			},
		},
		"removes obsolete path segments from base url": {
			args: []string{
				"--base-url",
				".//lsgo////",
			},
			expectedConfig: Config{
				addr:              "localhost:8080",
				folder:            "./public",
				baseURL:           "/lsgo/",
				maxInlineFileSize: 1_048_576,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given When
			var buf bytes.Buffer
			actualConfig, err := parseFlags(tc.args, &buf)

			// Then
			// when everything is fine nothing is written to STDERR
			require.NoError(t, err)
			assert.Empty(t, buf)
			assert.Equal(t, tc.expectedConfig, *actualConfig)
		})
	}
}

func Test_should_inform_about_parse_errors(t *testing.T) {
	testCases := map[string]struct {
		args           []string
		expectedError  string
		expectedOutput string
	}{
		"help text": {
			args:           []string{"--help"},
			expectedOutput: "Usage of lsgo",
			expectedError:  "help requested",
		},
		"max-inline-file-size not a number": {
			args:           []string{"--max-inline-file-size", "aa"},
			expectedOutput: "invalid value \"aa\" for flag -max-inline-file-size: parse error",
			expectedError:  "failed to parse config: invalid value \"aa\" for flag -max-inline-file-size",
		},
		"empty argument": {
			args:           []string{"--base-url"},
			expectedOutput: "flag needs an argument: -base-url",
			expectedError:  "failed to parse config: flag needs an argument: -base-url",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given When
			var buf bytes.Buffer
			_, err := parseFlags(tc.args, &buf)

			// Then
			require.ErrorContains(t, err, tc.expectedError)
			assert.Contains(t, buf.String(), tc.expectedOutput)
		})
	}
}
