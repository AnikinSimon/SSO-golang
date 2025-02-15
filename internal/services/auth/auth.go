package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrAppExists          = errors.New("app already exists")
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	appSaver     AppSaver
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaverUser(
		ctx context.Context,
		email string,
		passHash []byte,
		app_id string,
	) (string, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID string) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID string) (models.App, error)
}

type AppSaver interface {
	SaveApp(ctx context.Context, name string, secret string) (string, error)
}

// Create new entity of Auth
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	appSaver AppSaver,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		appSaver:     appSaver,
		log:          log,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID string,
) (string, error) {
	const op = "services.auth.Login"

	log := a.log.With(
		slog.String("op", op),
		// slog.String("email", email),
	)

	log.Info("login user")

	user, err := a.userProvider.User(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error:", err.Error()))

			return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to login user", slog.String("error:", err.Error()))

		return "", fmt.Errorf("%s %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Passhash, []byte(password)); err != nil {
		a.log.Info("Invalid credentials", slog.String("error:", err.Error()))

		return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)

	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("invalid app id", slog.String("error:", err.Error()))

			return "", fmt.Errorf("%s %w", op, ErrInvalidAppID)
		}
		return "", fmt.Errorf("%s %w", op, err)
	}

	log.Info("user logged in succesfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)

	if err != nil {
		a.log.Error("failed to generate token", slog.String("error:", err.Error()))

		return "", fmt.Errorf("%s %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
	app_id string,
) (string, error) {
	const op = "services.auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		// slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to generate password hash", slog.String("error:", err.Error()))

		return "", fmt.Errorf("%s %w", op, err)
	}

	id, err := a.userSaver.SaverUser(ctx, email, passHash, app_id)

	if err != nil {

		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("user already exists", slog.String("error:", err.Error()))

			return "", fmt.Errorf("%s %w", op, ErrUserExists)
		}

		log.Error("failed to save user", slog.String("error:", err.Error()))
		return "", fmt.Errorf("%s %w", op, err)
	}

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID string) (bool, error) {
	const op = "services.auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		// slog.String("email", email),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)

	if err != nil {
		return false, fmt.Errorf("%s %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) RegisterNewApp(
	ctx context.Context,
	name string,
	secret string,
) (string, error) {
	const op = "services.auth.RegisterNewApp"

	log := a.log.With(
		slog.String("op", op),
		// slog.String("email", email),
	)

	log.Info("registering app")

	id, err := a.appSaver.SaveApp(ctx, name, secret)

	if err != nil {

		if errors.Is(err, storage.ErrAppExists) {
			a.log.Warn("app already exists", slog.String("error:", err.Error()))

			return "", fmt.Errorf("%s %w", op, ErrAppExists)
		}

		log.Error("failed to save app", slog.String("error:", err.Error()))
		return "", fmt.Errorf("%s %w", op, err)
	}

	return id, nil
}
