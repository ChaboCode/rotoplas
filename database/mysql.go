package database

import (
	"database/sql"
	"fmt"
	"log"
	"rotoplas/models"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type File struct {
	ID        int64
	Name      string
	Size      int64
	UploadIp  string
	CreatedAt string
	MimeType  string
	Hidden    bool
}

func ConnectMySQL() {
	cfg := mysql.NewConfig()
	cfg.User = "root"
	cfg.Passwd = "sietemachete"
	cfg.Net = "tcp"
	cfg.Addr = "mysql:3306"
	// cfg.DBName = "rotoplas"

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS rotoplas")
	if err != nil {
		log.Fatal("Error creando base de datos:", err)
	}

	cfg.DBName = "rotoplas" // Establecer el nombre de la base de datos

	// Conectarse ahora a la base de datos creada
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal("Error reconectando a la base de datos:", err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Connected to the database successfully!")
	initTable()
}

func AddFile(file models.File) (int64, error) {
	result, err := db.Exec("INSERT INTO rotoplas (name, size, upload_ip, created_at, mime_type, hidden) VALUES (?, ?, ?, ?, ?, ?)",
		file.Name, file.Size, file.UploadIP, file.CreatedAt, file.MimeType, file.Hidden)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func DeleteFile(name string) error {
	_, err := db.Exec("DELETE FROM rotoplas WHERE name = ?", name)
	if err != nil {
		return err
	}
	return nil
}

func ListFiles(count int, page int) ([]File, error) {
	var files []File

	rows, err := db.Query("SELECT * FROM rotoplas WHERE hidden IS NOT TRUE ORDER BY created_at DESC LIMIT ? OFFSET ? ", count, (page-1)*count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file File
		if err := rows.Scan(&file.ID, &file.Name, &file.Size, &file.UploadIp, &file.CreatedAt, &file.MimeType, &file.Hidden); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func Count() (int64, error) {
	var count int64
	err := db.QueryRow("SELECT COUNT(*) FROM rotoplas WHERE hidden IS NOT TRUE").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func initTable() {
	_, err := db.Exec(`CREATE TABLE rotoplas (
  id int NOT NULL AUTO_INCREMENT,
  name varchar(100) NOT NULL,
  size int DEFAULT NULL,
  upload_ip char(15) DEFAULT NULL,
  created_at datetime DEFAULT NULL,
  mime_type varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  hidden bool DEFAULT NULL,
  PRIMARY KEY (id),
  KEY rotoplas_name_IDX (name) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`)

	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1050 {
			// Table already exists, ignore the error
			fmt.Println("Table 'rotoplas' already exists, skipping creation.")
		} else {
			log.Fatalf("Error creating table: %v", err)
		}
	}
}
