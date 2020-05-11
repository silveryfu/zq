package ingest_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/brimsec/zq/zqd/api"
	"github.com/brimsec/zq/zqd/ingest"
	"github.com/brimsec/zq/zqd/space"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempSpace(t *testing.T, spaces *space.Manager) *space.Space {
	sp, err := spaces.Create(t.Name(), "")
	require.NoError(t, err)

	space, err := spaces.Get(sp.ID())
	require.NoError(t, err)

	return space
}
func writeTempFile(t *testing.T, data string) string {
	f, err := ioutil.TempFile("", "testfile")
	require.NoError(t, err)
	name := f.Name()
	defer f.Close()
	_, err = f.WriteString(data)
	require.NoError(t, err)
	return name
}

func TestLogsErrInFlight(t *testing.T) {
	src := `
#0:record[_path:string,ts:time,uid:bstring]
0:[conn;1521911723.205187;CBrzd94qfowOqJwCHa;]
0:[conn;1521911721.255387;C8Tful1TvM3Zf5x8fl;]
`
	root, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(root)

	spaces := space.NewManager(root, nil)
	s := createTempSpace(t, spaces)
	ctx := context.Background()

	f := writeTempFile(t, src)

	errCh1 := make(chan error)
	errCh2 := make(chan error)
	go func() {
		_, err := ingest.NewLogTransaction(ctx, s.Storage, api.LogPostRequest{Paths: []string{f}})
		errCh1 <- err
	}()
	go func() {
		_, err := ingest.NewLogTransaction(ctx, s.Storage, api.LogPostRequest{Paths: []string{f}})
		errCh2 <- err
	}()
	err1 := <-errCh1
	err2 := <-errCh2
	if err1 == nil {
		assert.EqualError(t, err2, ingest.ErrIngestProcessInFlight.Error())
		return
	}
	if err2 == nil {
		assert.EqualError(t, err1, ingest.ErrIngestProcessInFlight.Error())
		return
	}
	assert.Fail(t, "expected only one error")
}
