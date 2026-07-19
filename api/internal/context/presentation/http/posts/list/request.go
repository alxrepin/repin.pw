package list

type Request struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
