package postgres

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/config"
	"sso/internal/domain/models"
	"sso/internal/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	UniqueConstraintEmail = "uni_users_email"
	UniqueConstraintApp   = "uni_apps_name"
)

func IsUniqueConstraintError(err error, constraintName string) bool {

	if pqErr, ok := err.(*pgconn.PgError); ok {
		return pqErr.Code == "23505" && pqErr.ConstraintName == constraintName
	}
	fmt.Println(err)
	return false
}

type Storage struct {
	db *gorm.DB
}

func getPostgresConn(cfg config.StorageConfig) string {
	return fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.User, cfg.Password, cfg.Database, cfg.Port)
}

func New(cfg config.StorageConfig) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := gorm.Open(postgres.Open(getPostgresConn(cfg)), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	err = db.AutoMigrate(&models.User{}, &models.App{})

	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaverUser(
	ctx context.Context,
	email string,
	passHash []byte,
	app_id string,
) (string, error) {
	const op = "storage.postgres.SaveUser"

	uid := uuid.New().String()

	user := models.User{ID: uid, Email: email, Passhash: passHash, AppID: app_id}

	tx := s.db.WithContext(ctx).Create(&user)

	if tx.Error != nil {
		if IsUniqueConstraintError(tx.Error, UniqueConstraintEmail) {
			return "", fmt.Errorf("%s %w", op, storage.ErrUserExists)
		}

		return "", fmt.Errorf("%s %w", op, tx.Error)
	}

	return uid, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.SaveUser"

	var user models.User
	tx := s.db.WithContext(ctx).Find(&user, "email = ?", email)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.User{}, fmt.Errorf("%s %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s %w", op, tx.Error)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, appID string) (models.App, error) {
	const op = "storage.postgres.App"

	var app models.App
	tx := s.db.WithContext(ctx).Find(&app, "ID = ?", appID)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.App{}, fmt.Errorf("%s %w", op, storage.ErrUserNotFound)
		}

		return models.App{}, fmt.Errorf("%s %w", op, tx.Error)
	}

	return app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, email string) (bool, error) {
	const op = "storage.postgres.SaveUser"

	var user models.User
	tx := s.db.WithContext(ctx).Find(&user, "email = ?", email)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("%s %w", op, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s %w", op, tx.Error)
	}

	return user.IsAdmin, nil
}

func (s *Storage) SaveApp(
	ctx context.Context,
	name string,
	secret string,
) (string, error) {
	const op = "storage.postgres.SaveApp"

	appId := uuid.New().String()

	app := models.App{ID: appId, Name: name, Secret: secret}

	tx := s.db.WithContext(ctx).Create(&app)

	if tx.Error != nil {
		fmt.Println(tx.Error)
		if IsUniqueConstraintError(tx.Error, UniqueConstraintApp) {
			return "", fmt.Errorf("%s %w", op, storage.ErrAppExists)
		}

		return "", fmt.Errorf("%s %w", op, tx.Error)
	}

	return appId, nil
}
