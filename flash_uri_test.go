package flash

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURI(t *testing.T) {
	var tt = []struct {
		fileConfig FileConfig
	}{
		{
			fileConfig: FileConfig{
				Path: "/this/is/absolute",
			},
		},
		{
			fileConfig: FileConfig{
				Path: "../this/is/relative",
			},
		},
		{
			fileConfig: FileConfig{
				Path: "./this/is/relative",
			},
		},
		{
			fileConfig: FileConfig{
				Path: "C:\\windows\\path",
			},
		},
		{
			fileConfig: FileConfig{
				Path: "\\windows\\path",
			},
		},
	}

	for _, tc := range tt {
		u, err := url.ParseRequestURI(tc.fileConfig.sinkURI())
		require.NoError(t, err)

		p := pathFromURI(u)

		assert.Equal(t, tc.fileConfig.Path, p)
	}
}
