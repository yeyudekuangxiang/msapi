package neteasy

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestArtist(t *testing.T) {
	client := NewClient("http://nas.znil.cn:3000")
	list, err := client.SearchMusic("周杰伦", 10, 0)
	require.Equal(t, nil, err)
	require.Greater(t, len(list), 0)
	t.Logf("list: %v", list)
}
func TestPlayUrl(t *testing.T) {
	client := NewClient("http://nas.znil.cn:3000")
	list, err := client.GetPlayUrl([]string{"509781655"})
	require.Equal(t, nil, err)
	require.Greater(t, len(list), 0)
	t.Logf("list: %v", list)
}
