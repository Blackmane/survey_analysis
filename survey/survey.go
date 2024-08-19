package survey

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var dataTableName = "survey_data"
var origTableName = "survey_data_orig"
var viewTableName = "survey_views"
var tagTableName = "survey_tag"
var dbPath = "survey.db"

func loadDB() error {
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func createDB(filePath string) error {
	responses, err := LoadCSV(filePath)
	if err != nil {
		return fmt.Errorf("error loading CSV: %w", err)
	}
	log.Println("Loaded responses:", responses)

	err = DeleteDBIfExists()
	if err != nil {
		return fmt.Errorf("error deleting old database: %w", err)
	}

	err = SaveToDB(responses)
	if err != nil {
		return fmt.Errorf("error saving to database: %w", err)
	}
	return nil
}

func LoadCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read csv data: %w", err)
	}

	return records, nil
}

func DeleteDBIfExists() error {
	if _, err := os.Stat(dbPath); err == nil {
		err := os.Remove(dbPath)
		if err != nil {
			return fmt.Errorf("could not delete existing database: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("could not check if database exists: %w", err)
	}
	return nil
}

func SaveToDB(records [][]string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	// Assuming the first row contains column names
	headers := records[0]
	createTableQuery := " (id INTEGER PRIMARY KEY,"
	for i, header := range headers {
		createTableQuery += fmt.Sprintf("%s TEXT", header)
		if i < len(headers)-1 {
			createTableQuery += ", "
		}
	}
	createTableQuery += ");"

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + dataTableName + createTableQuery)
	if err != nil {
		return fmt.Errorf("could not create table: %w", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + origTableName + createTableQuery)
	if err != nil {
		return fmt.Errorf("could not create table: %w", err)
	}

	insertQuery := " ("
	for i, header := range headers {
		insertQuery += header
		if i < len(headers)-1 {
			insertQuery += ", "
		}
	}
	insertQuery += ") VALUES ("
	for i := range headers {
		insertQuery += "?"
		if i < len(headers)-1 {
			insertQuery += ", "
		}
	}
	insertQuery += ");"

	for i, record := range records[1:] {
		_, err = db.Exec("INSERT INTO "+dataTableName+insertQuery, toInterfaceSlice(i, record)...)
		if err != nil {
			return fmt.Errorf("could not insert record: %w", err)
		}
		_, err = db.Exec("INSERT INTO "+origTableName+insertQuery, toInterfaceSlice(i, record)...)
		if err != nil {
			return fmt.Errorf("could not insert record: %w", err)
		}
	}

	createViewTableQuery := `
        CREATE TABLE IF NOT EXISTS survey_views (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            view_name TEXT,
            columns TEXT
        );`
	_, err = db.Exec(createViewTableQuery)
	if err != nil {
		return fmt.Errorf("could not create views table: %w", err)
	}


	return nil
}
func toInterfaceSlice(i int, strs []string) []interface{} {
	result := make([]interface{}, len(strs)+1)
	result[0] = i
	for i, v := range strs {
		result[i+1] = v
	}
	return result
}

func SaveView(dbPath, viewName string, columns string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO survey_views (view_name, columns) VALUES (?, ?)", viewName, columns)
	if err != nil {
		return fmt.Errorf("could not insert view: %w", err)
	}

	return nil
}

func GetViews() (map[string][]string, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT view_name, columns FROM survey_views")
	if err != nil {
		return nil, fmt.Errorf("could not query views: %w", err)
	}
	defer rows.Close()

	views := make(map[string][]string)
	for rows.Next() {
		var viewName string
		var columnsStr string
		err := rows.Scan(&viewName, &columnsStr)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}

		var columns []string
		fmt.Sscanf(columnsStr, "%v", &columns)
		views[viewName] = columns
	}

	return views, nil
}

func getQuery(query string) (*sql.Rows, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	return rows, nil
}

func getData() (*sql.Rows, error) {
	return getQuery("SELECT * FROM " + dataTableName)
}

func getViewList() (*sql.Rows, error) {
	return getQuery("SELECT view_name FROM survey_views")
}

func getViewUi() (*sql.Rows, error) {
	return getQuery("SELECT * FROM " + dataTableName + " LIMIT 1")
}

func getViewData(viewName string) (*sql.Rows, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()
	var columns string
	err = db.QueryRow("SELECT columns FROM survey_views WHERE view_name = ?", viewName).Scan(&columns)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}

	selectedCols := strings.Split(columns, ",")
	query := "SELECT " + strings.Join(selectedCols, ", ") + " FROM " + dataTableName
	return db.Query(query)
}
