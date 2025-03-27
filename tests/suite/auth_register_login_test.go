package suite

import (
	"testing"
	"time"

	ssov1 "sso/streaming/go/sso"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppId = ""
	appId      = ""
	appSecret  = "test-secret"

	passDefeaultLen = 10
)

func TestRegister_Login_HappyPath(t *testing.T) {
	ctx, st := New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	appName := gofakeit.Name()
	appSecret := randomFakePassword()

	registerAppResponse, err := st.AuthClient.RegisterApp(ctx, &ssov1.RegisterAppRequest{
		Name:   appName,
		Secret: appSecret,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerAppResponse.GetAppUuid())

	registerResponse, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
		AppUuid:  registerAppResponse.GetAppUuid(),
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerResponse.GetUserUuid())

	loginResponse, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppUuid:  registerAppResponse.GetAppUuid(),
	})

	require.NoError(t, err)

	loginTime := time.Now()

	token := loginResponse.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})

	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, registerResponse.GetUserUuid(), claims["uid"].(string))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, registerAppResponse.GetAppUuid(), claims["app_id"].(string))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)

}

func TestRegister_Login_DuplicatedRegistration(t *testing.T) {
	ctx, st := New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	appName := gofakeit.Name()
	appSecret := randomFakePassword()

	registerAppResponse, err := st.AuthClient.RegisterApp(ctx, &ssov1.RegisterAppRequest{
		Name:   appName,
		Secret: appSecret,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerAppResponse.GetAppUuid())

	registerResponse, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
		AppUuid:  registerAppResponse.GetAppUuid(),
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerResponse.GetUserUuid())

	registerResponse, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
		AppUuid:  registerAppResponse.GetAppUuid(),
	})

	require.Error(t, err)
	assert.Empty(t, registerResponse.GetUserUuid())
	assert.ErrorContains(t, err, "user already exists")

}

func TestRegisterApp_DuplicatedRegistration(t *testing.T) {
	ctx, st := New(t)

	appName := gofakeit.Name()
	appSecret := randomFakePassword()

	registerAppResponse, err := st.AuthClient.RegisterApp(ctx, &ssov1.RegisterAppRequest{
		Name:   appName,
		Secret: appSecret,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerAppResponse.GetAppUuid())

	registerAppResponse, err = st.AuthClient.RegisterApp(ctx, &ssov1.RegisterAppRequest{
		Name:   appName,
		Secret: appSecret,
	})

	require.Error(t, err)
	assert.Empty(t, registerAppResponse.GetAppUuid())
	assert.ErrorContains(t, err, "app already exists")

}

func TestRegister_FailCase(t *testing.T) {
	ctx, st := New(t)

	appName := gofakeit.Name()
	appSecret := randomFakePassword()

	registerAppResponse, err := st.AuthClient.RegisterApp(ctx, &ssov1.RegisterAppRequest{
		Name:   appName,
		Secret: appSecret,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerAppResponse.GetAppUuid())

	tests := []struct {
		name        string
		email       string
		password    string
		appId       string
		expectedErr string
	}{
		{
			name:        "Register with empty password",
			email:       gofakeit.Email(),
			password:    "",
			appId:       registerAppResponse.GetAppUuid(),
			expectedErr: "password is required",
		},
		{
			name:        "Register with empty email",
			email:       "",
			password:    randomFakePassword(),
			appId:       registerAppResponse.GetAppUuid(),
			expectedErr: "email is required",
		},

		{
			name:        "Register with empty app_id",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appId:       registerAppResponse.GetAppUuid(),
			expectedErr: "app_id is required",
		},

		{
			name:        "Register with empty params",
			email:       "",
			password:    "",
			appId:       "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
				AppUuid:  appId,
			})

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestRegisterApp_FailCase(t *testing.T) {
	ctx, st := New(t)

	tests := []struct {
		name        string
		appName     string
		appSecret   string
		expectedErr string
	}{
		{
			name:        "Register with empty secret",
			appName:     gofakeit.Name(),
			appSecret:   "",
			expectedErr: "secret is required",
		},
		{
			name:        "Register with empty name",
			appName:     "",
			appSecret:   randomFakePassword(),
			expectedErr: "name is required",
		},

		{
			name:        "Register with both empty",
			appName:     "",
			appSecret:   "",
			expectedErr: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.RegisterApp(ctx, &ssov1.RegisterAppRequest{
				Name:   tt.appName,
				Secret: tt.appSecret,
			})

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, true, passDefeaultLen)
}
