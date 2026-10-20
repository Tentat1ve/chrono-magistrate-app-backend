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

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("ошибка подключения к БД: %v", err)
	}

	// Создание таблиц users, dignitarie_lists, dignitarie_list_likes
	err = db.AutoMigrate(
		&ds.User{},
		&ds.DignitarieList{},
		&ds.DignitarieListLike{},
	)
	if err != nil {
		logrus.Fatalf("ошибка миграции: %v", err)
	}

	// Дата окончания не раньше даты начала
	err = db.Exec(`DO $$ BEGIN
		ALTER TABLE dignitarie_lists ADD CONSTRAINT chk_dignitarie_lists_dates
			CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date);
		EXCEPTION WHEN duplicate_object THEN NULL;
	END $$;`).Error
	if err != nil {
		logrus.Fatalf("ошибка добавления ограничения: %v", err)
	}

	logrus.Info("миграция выполнена")
}
