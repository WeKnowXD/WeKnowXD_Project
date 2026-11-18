package search

import (
	"database/sql"
	"strings"

	"github.com/WeKnowXD/WeKnowXD_Project/internal/models"
)

const snippetLength = 800

func Get(db *sql.DB, q string, language string) ([]models.SearchResult, error) {
	if q == "" {
		return []models.SearchResult{}, nil
	}
	if language == "" {
		language = "en"
	}

	rows, err := db.Query(
		"SELECT title, url, substr(content, 1, ?) FROM pages WHERE language = ? AND content LIKE ?",
		snippetLength+1, language, "%"+q+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []models.SearchResult{}
	for rows.Next() {
		var res models.SearchResult
		if err := rows.Scan(&res.Title, &res.URL, &res.Content); err != nil {
			return nil, err
		}
		res.Content = MakeSnippet(res.Content, snippetLength)
		results = append(results, res)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func MakeSnippet(text string, max int) string {
	if len(text) <= max {
		return text
	}
	cut := text[:max]
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return strings.TrimSpace(cut) + "..."
}
