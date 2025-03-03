package cli

import (
	"context"
	"errors"
	"net/url"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Login descriptions
const (
	LoginUsageText            = "Logs you in"
	LoginDescription          = "Log in to NordVPN by using the default method. We'll take you to your browser for login and then bring you back to the app. Other login methods are available as options."
	LoginNordAccountUsageText = "This option is no longer available."
	LoginFlagTokenUsageText   = "Log in to NordVPN by using a token generated in your Nord Account. This login option doesn't support multi-factor authentication. Tokens are revoked at logout. Use \"nordvpn logout --help\" for more info." // #nosec
	LoginCallbackUsageText    = "Complete the login manually if your browser fails to open the app. After you successfully log in on your browser, copy the link of the \"Continue\" button and paste it enclosed in quotation marks as an argument for this option."
)

func (c *cmd) Login(ctx *cli.Context) error {
	resp, err := c.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil || resp.GetValue() {
		return formatError(internal.ErrAlreadyLoggedIn)
	}

	if ctx.IsSet(flagLoginCallback) {
		return c.oauth2(ctx, true)
	}

	if ctx.IsSet(flagToken) {
		err = c.loginWithToken(ctx)
		if err != nil {
			return formatError(err)
		}

		return nil
	}

	return c.login(pb.LoginType_LoginType_LOGIN)
}

func (c *cmd) login(requestType pb.LoginType) error {
	resp, err := c.client.LoginOAuth2(
		context.Background(),
		&pb.LoginOAuth2Request{
			Type: requestType,
		},
	)
	if err != nil {
		return formatError(err)
	}

	switch resp.Status {
	case pb.LoginOAuth2Status_UNKNOWN_OAUTH2_ERROR:
		return formatError(internal.ErrUnhandled)
	case pb.LoginOAuth2Status_NO_NET:
		return formatError(internal.ErrNoNetWhenLoggingIn)
	case pb.LoginOAuth2Status_ALREADY_LOGGED_IN:
		return formatError(internal.ErrAlreadyLoggedIn)
	case pb.LoginOAuth2Status_SUCCESS:
		if url := resp.Ulr; url != "" {
			color.Green("Continue in the browser: %s", url)
		} else {
			return formatError(internal.ErrUnhandled)
		}
	}

	return nil
}

func (c *cmd) loginWithToken(ctx *cli.Context) error {
	// nordvpn login --token b50fc06c2bf6331522c1ef5f1d449ca99b818a16ef10253d67b4a4804d9x0xd6
	token := ctx.Args().First()
	if token == "" {
		return formatError(errors.New(client.TokenLoginFailure))
	}

	resp, err := c.client.LoginWithToken(context.Background(), &pb.LoginWithTokenRequest{
		Token: token,
	})
	if err != nil {
		return formatError(err)
	}
	return LoginRespHandler(ctx, resp)
}

func LoginRespHandler(ctx *cli.Context, resp *pb.LoginResponse) error {
	switch resp.Type {
	case internal.CodeGatewayError:
		return formatError(errors.New(client.ConnectTimeoutError))
	case internal.CodeUnauthorized:
		return formatError(errors.New(client.LegacyLoginFailure))
	case internal.CodeBadRequest:
		return formatError(errors.New(client.LoginFailure))
	case internal.CodeTokenLoginFailure:
		return formatError(errors.New(client.TokenLoginFailure))
	case internal.CodeTokenInvalid:
		return formatError(errors.New(client.TokenInvalid))
	case internal.CodeSuccess:
		color.Green(LoginSuccess, ctx.App.Name)
		color.Yellow("\nNOTE: %s", MsgNordVPNGroup)
	}
	return nil
}

// oauth2 is called by the browser during login via OAuth2.
func (c *cmd) oauth2(ctx *cli.Context, regularLogin bool) error {
	if ctx.NArg() != 1 {
		return formatError(errors.New("expected a url"))
	}

	url, err := url.Parse(ctx.Args().First())
	if err != nil {
		return formatError(err)
	}

	if url.Scheme != "nordvpn" {
		return formatError(errors.New("expected a url with nordvpn scheme"))
	}

	loginType := pb.LoginType_LoginType_LOGIN
	if !regularLogin {
		loginType = pb.LoginType_LoginType_SIGNUP
	}

	_, err = c.client.LoginOAuth2Callback(context.Background(), &pb.LoginOAuth2CallbackRequest{
		Token: url.Query().Get("exchange_token"),
		Type:  loginType,
	})
	if err != nil {
		return formatError(err)
	}

	color.Green(LoginSuccess, ctx.App.Name)
	color.Yellow("\nNOTE: %s", MsgNordVPNGroup)
	return nil
}
