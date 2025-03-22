package utils

import (
	"log"
	"os"
	"path/filepath"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(databaseURL string) {
	// Получаем текущую рабочую директорию
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Ошибка при получении рабочей директории: %v", err)
	}
	log.Printf("Текущая рабочая директория: %s", wd)

	// Формируем абсолютный путь до папки с миграциями
	migrationsPath := filepath.Join(wd, "db/migrations")

	// Проверяем, существует ли папка с миграциями
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Fatalf("Папка с миграциями не найдена: %s", migrationsPath)
	}

	// Применяем миграции
	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Ошибка при создании миграции: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка при применении миграций: %v", err)
	}

	log.Println("Миграции успешно применены")
}
