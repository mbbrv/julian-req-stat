package dbjulian

import (
	"github.com/jmoiron/sqlx"
	"strconv"
	"time"
)

type Article struct {
	ID           int64     `db:"id"`
	CategoryID   int64     `db:"category_id"`
	ItemWBName   string    `db:"item_wb_name"`
	ProductName  string    `db:"product_name"`
	Barcode      string    `db:"barcode"`
	Nomenclature string    `db:"nomenclature"`
	Photo        string    `db:"photo"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Category struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type QuantityStat struct {
	ID        int64     `db:"id"`
	ArticleID int64     `db:"article_id"`
	Quantity  int       `db:"quantity"`
	Date      time.Time `db:"date"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RequestStat struct {
	ID          int64     `db:"id"`
	ArticleID   int64     `db:"article_id"`
	Name        string    `db:"name"`
	Results     string    `db:"results"`
	FrequencyWB string    `db:"frequency_wb"`
	SearchPlace string    `db:"search_place"`
	Date        time.Time `db:"date"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type SalesStat struct {
	ID        int64     `db:"id"`
	ArticleID int64     `db:"article_id"`
	Sales     int       `db:"sales"`
	Date      time.Time `db:"date"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func GetArticles(db *sqlx.DB) []Article {
	var articlesDB []Article
	db.Select(&articlesDB, "SELECT * FROM articles")

	return articlesDB
}

func GetNomenclatureFromArticles(articles []Article) []string {
	nomenclature := make([]string, len(articles))

	for i := range articles {
		nomenclature[i] = articles[i].Nomenclature
	}

	return nomenclature
}

func IsRequestStatExists(db *sqlx.DB, articleId int64, date time.Time, name string) (bool, error) {
	query := "SELECT * FROM requests_stat WHERE article_id = :article_id AND date = :date AND name = :name;"
	rows, err := db.NamedExec(query, map[string]interface{}{
		"article_id": articleId,
		"date":       date,
		"name":       name,
	})
	if err != nil {
		return false, err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func GetCheckRequestStatSlice(db *sqlx.DB) (map[string]bool, error) {
	var requestStat []RequestStat
	res := make(map[string]bool)

	err := db.Select(&requestStat, "SELECT * FROM requests_stat")
	if err != nil {
		return nil, err
	}

	for i := range requestStat {
		artId := strconv.FormatInt(requestStat[i].ArticleID, 10)
		date := requestStat[i].Date.Format("2006-01-02")
		name := requestStat[i].Name

		res[artId+date+name] = true
	}

	return res, nil
}

func InsertRequestStat(db *sqlx.DB, requestStat []RequestStat) (bool, error) {
	query := "INSERT INTO requests_stat (article_id, name, results, frequency_wb, search_place, date, created_at, updated_at) VALUES (:article_id, :name, :results, :frequency_wb, :search_place, :date, :created_at, :updated_at);"
	rows, err := db.NamedExec(query, requestStat)
	if err != nil {
		return false, err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
