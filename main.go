package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "validator"
	password = "val1dat0r"
	dbname   = "project-sem-1"
)

var db *sql.DB

func initDB() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка проверки соединения: %v", err)
	}

	//fmt.Println("Успешное подключение к базе данных!")
}

func pricesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleZipRequest(w, r)
	case http.MethodGet:
		getPricesHandler(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleZipRequest(w http.ResponseWriter, r *http.Request) {
	// Чтение zip-архива из тела запроса
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Разархивация zip-файла
	archive, err := zip.NewReader(file, r.ContentLength)
	if err != nil {
		http.Error(w, "Ошибка при разархивировании файла", http.StatusInternalServerError)
		return
	}

	// Инициализация счетчиков
	var totalItems, totalCategories int
	var totalPrice float64
	categorySet := make(map[string]bool)

	// Проходим по всем файлам в архиве
	for _, zipFile := range archive.File {
		// Проверяем, что файл имеет расширение .csv
		if strings.HasSuffix(zipFile.Name, ".csv") {
			zipFileReader, err := zipFile.Open()
			if err != nil {
				http.Error(w, "Ошибка при открытии файла из архива", http.StatusInternalServerError)
				return
			}
			defer zipFileReader.Close()

			// Читаем CSV из архива
			csvReader := csv.NewReader(zipFileReader)
			firstLine := true // Переменная для проверки заголовка
			for {
				record, err := csvReader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					http.Error(w, "Ошибка при чтении CSV файла", http.StatusInternalServerError)
					return
				}

				// Пропускаем заголовок
				if firstLine {
					firstLine = false
					continue
				}

				if len(record) < 5 {
					http.Error(w, "Неверный формат данных в CSV", http.StatusBadRequest)
					return
				}

				id := record[0]
				name := record[1]
				category := record[2]
				price := record[3]
				createDate := record[4]

				// Конвертируем цену в float64
				priceValue, err := strconv.ParseFloat(price, 64)
				if err != nil {
					log.Printf("Ошибка преобразования цены: %v", err)
					http.Error(w, "Ошибка преобразования цены", http.StatusBadRequest)
					return
				}

				// Преобразуем createDate в формат DATE
				parsedDate, err := time.Parse("2006-01-02", createDate)
				if err != nil {
					http.Error(w, "Ошибка преобразования даты", http.StatusBadRequest)
					return
				}

				// Добавляем данные в базу данных
				query := `INSERT INTO prices (id, name, category, price, created_at) 
						VALUES ($1, $2, $3, $4, $5) 
						ON CONFLICT (id) DO NOTHING`
				_, err = db.Exec(query, id, name, category, priceValue, parsedDate)
				if err != nil {
					log.Printf("Ошибка преобразования цены: %v", err)
					http.Error(w, "Ошибка при добавлении данных в базу", http.StatusInternalServerError)
					return
				}
				log.Printf("Добавлена запись: %s, %s, %.2f", name, category, priceValue)

				// Обновляем счетчики
				totalItems++
				if !categorySet[category] {
					categorySet[category] = true
					totalCategories++
				}
				totalPrice += priceValue
			}
		}
	}

	// Формируем ответ
	response := map[string]interface{}{
		"total_items":      totalItems,
		"total_categories": totalCategories,
		"total_price":      totalPrice,
	}

	// Отправляем JSON с результатами
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getPricesHandler обрабатывает GET-запрос для получения всех цен
func getPricesHandler(w http.ResponseWriter, r *http.Request) {
	// Создаем временный файл для zip
	tmpZipFile, err := os.CreateTemp("", "prices-*.zip")
	if err != nil {
		http.Error(w, "Ошибка создания временного файла", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpZipFile.Name())

	// Открываем zip-архив для записи
	zipWriter := zip.NewWriter(tmpZipFile)

	// Создаем CSV-файл внутри архива
	csvFile, err := zipWriter.Create("data.csv")
	if err != nil {
		http.Error(w, "Ошибка создания файла в архиве", http.StatusInternalServerError)
		return
	}

	// Пишем данные в CSV-файл
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Запрос данных из базы данных
	rows, err := db.Query(`SELECT name, category, price FROM prices`)
	if err != nil {
		http.Error(w, "Ошибка извлечения данных из базы", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Запись заголовков CSV
	writer.Write([]string{"Name", "Category", "Price"})

	//log.Println("Извлечение данных из базы начато")

	for rows.Next() {
		var name, category string
		var price float64
		if err := rows.Scan(&name, &category, &price); err != nil {
			log.Printf("Ошибка обработки строки: %v", err)
			http.Error(w, "Ошибка обработки строки", http.StatusInternalServerError)
			return
		}
		//log.Printf("Извлечена запись: %s, %s, %.2f", name, category, price)
		writer.Write([]string{name, category, strconv.FormatFloat(price, 'f', 2, 64)})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Printf("Ошибка записи в CSV: %v", err)
		http.Error(w, "Ошибка записи в CSV", http.StatusInternalServerError)
		return
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Ошибка закрытия архива", http.StatusInternalServerError)
		return
	}

	// Отправка архива в ответ
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="prices.zip"`)
	http.ServeFile(w, r, tmpZipFile.Name())
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/api/v0/prices", pricesHandler)

	//fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
