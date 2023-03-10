package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestFileshareAccept(t *testing.T) {
	app := &cli.App{
		Name: "fileshare",
		Commands: []*cli.Command{
			fileshareAcceptCommand(&cmd{
				loaderInterceptor: &LoaderInterceptor{},
			}),
		},
	}

	err := app.Run([]string{"fileshare", "accept", "--path", "/tmp/nosuchdir", "someid"})
	assert.Error(t, err)
}
