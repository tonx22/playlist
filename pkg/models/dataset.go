package models

type Song struct {
	Id          int    `json:"id"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	Prev        int    `json:"prev"`
	Next        int    `json:"next"`
}

type ResponseError struct {
	ErrorDescr string
	Status     int
}

func (c ResponseError) Error() string {
	return c.ErrorDescr
}
