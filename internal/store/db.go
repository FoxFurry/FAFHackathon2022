package store

import (
	"context"
	"fmt"
	"log"

	"fafhackathon2022/internal/store/models"
	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type Store interface {
	GetUserData(ctx context.Context, userUUID string) (*models.User, error)
	CreateUser(ctx context.Context, userUUID string) (*models.User, error)

	GetNearestNCoordinates(ctx context.Context, limit int, radius float64, location redis.GeoLocation) ([]redis.GeoLocation, error)
	UpdateUserCoordinates(ctx context.Context, location redis.GeoLocation) error
	GetUserCoordinates(ctx context.Context, userUUID string) (*redis.GeoLocation, error)
	UpdateUserRole(ctx context.Context, userUUID string, userRole models.Role) error
	DeleteUserRole(ctx context.Context, userUUID string) error
	GetUserRole(ctx context.Context, userUUID string) (models.Role, error)
}

type store struct {
	sql   *sqlx.DB
	redis *redis.Client
}

func CreateDB() Store {
	var (
		d      store
		dbHost = viper.GetString("DATABASE_HOST")
		dbName = viper.GetString("DATABASE_NAME")
		dbPort = viper.GetString("DATABASE_PORT")
		dbUser = viper.GetString("DATABASE_USER")
		dbPass = viper.GetString("DATABASE_PASSWORD")

		redisAddr = viper.GetString("REDIS_ADDR")
		redisPass = viper.GetString("REDIS_PASS")

		err error
	)

	if dbHost == "" ||
		dbName == "" ||
		dbPort == "" ||
		dbUser == "" ||
		dbPass == "" ||
		redisAddr == "" {
		log.Panicln("Could not open DB connection. Some of the database env are missing")
	}

	cfg := mysql.Config{
		User:      dbUser,
		Passwd:    dbPass,
		Addr:      fmt.Sprintf("%s:%s", dbHost, dbPort),
		DBName:    dbName,
		ParseTime: true,
	}

	d.sql, err = sqlx.Connect("mysql", cfg.FormatDSN())
	if err != nil {
		log.Panicf("Could not open SQL connection: %v", err)
	}

	d.redis = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	})

	d.redis.FlushDB(context.Background())

	return &d
}
