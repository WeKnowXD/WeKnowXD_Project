package models

type PageData struct {
	Query   string
	Results []struct {
		Title       string
		URL         string
		Description string
	}
	Flashes []string
	User    *SessionData
}

type RegisterData struct {
	Error    string
	Username string
	Email    string
}

type LoginData struct {
	Error    string
	Username string
}
