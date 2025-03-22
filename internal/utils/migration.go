package utils

import (
	"fmt"
	"runtime"
	"log"
	"os"
	"path/filepath"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// getRepoRoot возвращает путь до корня репозитория.
func getRepoRoot() (string, error) {
	// Получаем путь до текущего файла
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("не удалось получить путь до текущего файла")
	}

	// Поднимаемся до корня репозитория
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	return repoRoot, nil
}

func RunMigrations(databaseURL string) {
	// Получаем путь до корня репозитория
	repoRoot, err := getRepoRoot()
	if err != nil {
		log.Fatalf("Ошибка при получении пути до корня репозитория: %v", err)
	}

	// Формируем абсолютный путь до папки с миграциями
	migrationsPath := filepath.Join(repoRoot, "internal/db/migrations")
	log.Printf("Путь до миграций: %s", migrationsPath)

	// Проверяем, существует ли папка с миграциями
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Fatalf("Папка с миграций не найдена: %s", migrationsPath)
	}

	// Применяем миграции
	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Ошибка при создании миграции: %v", err)
	}

	// Применяем миграции
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Миграции уже применены")
		} else {
			log.Fatalf("Ошибка при применении миграций: %v", err)
		}
	} else {
		log.Println("Миграции успешно применены")
	}
}


// func RunMigrations(databaseURL string) {
// 	// Получаем текущую рабочую директорию
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		log.Fatalf("Ошибка при получении рабочей директории: %v", err)
// 		fmt.Printf("Ошибка при получении рабочей директории: %v", err)
// 	}
// 	log.Printf("Текущая рабочая директория: %s", wd)

// 	// Формируем абсолютный путь до папки с миграциями
// 	migrationsPath := filepath.Join(wd, "../../internal/db/migrations")

// 	// Проверяем, существует ли папка с миграциями
// 	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
// 		log.Fatalf("Папка с миграциями не найдена: %s", migrationsPath)
// 		fmt.Printf("Папка с миграциями не найдена: %s", migrationsPath)
// 	}

// 	// Применяем миграции
// 	m, err := migrate.New(
// 		"file://"+migrationsPath,
// 		databaseURL,
// 	)
// 	if err != nil {
// 		log.Fatalf("Ошибка при создании миграции: %v", err)
// 		fmt.Printf("Ошибка при создании миграции: %v", err)
// 	}

// 	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
// 		log.Fatalf("Ошибка при применении миграций: %v", err)
// 		fmt.Printf("Ошибка при применении миграций: %v", err)
// 	}

// 	log.Println("Миграции успешно применены")
// 	fmt.Printf("Миграции успешно применены")
// }
