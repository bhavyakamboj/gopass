package action

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestFind(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	app := cli.NewApp()

	actName := "action.test"

	if runtime.GOOS == "windows" {
		actName = "action.test.exe"
	}

	// find
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	c.Context = ctx
	if err := act.Find(c); err == nil || err.Error() != fmt.Sprintf("Usage: %s find <NEEDLE>", actName) {
		t.Errorf("Should fail: %s", err)
	}

	// find fo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"fo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Find(c))
	assert.Equal(t, "Found exact match in 'foo'\nsecret", strings.TrimSpace(buf.String()))
	buf.Reset()

	// testing the safecontent case
	ctx = ctxutil.WithShowSafeContent(ctx, true)
	c.Context = ctx
	assert.Error(t, act.Find(c))
	buf.Reset()

	// testing with the clip flag set
	bf := cli.BoolFlag{
		Name:  "clip",
		Usage: "clip",
	}
	assert.NoError(t, bf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"-clip", "fo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Find(c))
	out := strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in 'foo'")
	buf.Reset()

	// safecontent case with force flag set
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"-force", "fo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Find(c))
	out = strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in 'foo'\nsecret")
	buf.Reset()

	// stopping with the safecontent tests
	ctx = ctxutil.WithShowSafeContent(ctx, false)

	// find yo
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"yo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.Error(t, act.Find(c))
	buf.Reset()

	// add some secrets
	assert.NoError(t, act.Store.Set(ctx, "bar/baz", secret.New("foo", "bar")))
	assert.NoError(t, act.Store.Set(ctx, "bar/zab", secret.New("foo", "bar")))
	buf.Reset()

	// find bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Find(c))
	assert.Equal(t, "bar/baz\nbar/zab", strings.TrimSpace(buf.String()))
	buf.Reset()
}
