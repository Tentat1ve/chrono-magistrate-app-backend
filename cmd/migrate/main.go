package main

import (
	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/dsn"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	logrus.Info("миграция схемы БД")

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("ошибка подключения к БД: %v", err)
	}

	// Создание таблиц users, dignitaries, dignitary_likes
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Dignitary{},
		&ds.DignitaryLike{},
	)
	if err != nil {
		logrus.Fatalf("ошибка миграции: %v", err)
	}

	// Год окончания не раньше года начала
	err = db.Exec(`DO $$ BEGIN
		ALTER TABLE dignitaries ADD CONSTRAINT chk_dignitaries_office_years
			CHECK (office_start IS NULL OR office_end IS NULL OR office_start <= office_end);
		EXCEPTION WHEN duplicate_object THEN NULL;
	END $$;`).Error
	if err != nil {
		logrus.Fatalf("ошибка добавления ограничения: %v", err)
	}

	logrus.Info("миграция выполнена")
}
