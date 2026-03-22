// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zapio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestCarriageReturnEdgeCases tests various edge cases for carriage return handling
func TestCarriageReturnEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "Multiple consecutive carriage returns",
			input:  "line1\r\rline2\r",
			expect: []string{"line1", "", "line2"},
		},
		{
			name:   "Carriage return at beginning",
			input:  "\rline1\rline2",
			expect: []string{"", "line1", "line2"},
		},
		{
			name:   "Carriage return at end without newline",
			input:  "line1\rline2\r",
			expect: []string{"line1", "line2"},
		},
		{
			name:   "Mixed separators complex",
			input:  "line1\n\rline2\r\nline3\n\rline4",
			expect: []string{"line1", "", "line2", "line3", "", "line4"},
		},
		{
			name:   "Progress updates with partial writes",
			input:  "10%\r20%\r30%\r40%\r",
			expect: []string{"10%", "20%", "30%", "40%"},
		},
		{
			name:   "Windows line endings mixed with Unix",
			input:  "line1\r\nline2\nline3\r\n",
			expect: []string{"line1", "line2", "line3"},
		},
		{
			name:   "Empty input with separators",
			input:  "\r\n\r\n",
			expect: []string{"", ""},
		},
		{
			name:   "Single carriage return",
			input:  "\r",
			expect: []string{""},
		},
		{
			name:   "Single Windows line ending",
			input:  "\r\n",
			expect: []string{""},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			core, observed := observer.New(zapcore.InfoLevel)
			writer := &Writer{
				Log:   zap.New(core),
				Level: zapcore.InfoLevel,
			}

			_, err := writer.Write([]byte(tt.input))
			require.NoError(t, err, "Write failed")

			require.NoError(t, writer.Close(), "Close failed")

			entries := observed.AllUntimed()
			assert.Equal(t, len(tt.expect), len(entries), "Number of logged entries mismatch")

			for i, expected := range tt.expect {
				if i < len(entries) {
					assert.Equal(t, expected, entries[i].Entry.Message,
						"Entry %d message mismatch", i)
				}
			}
		})
	}
}
