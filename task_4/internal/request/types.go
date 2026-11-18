package request

import "strings"

type Register struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,request_email,maxbytes=254"`
	Password string `json:"password" validate:"required,min=8,maxbytes=72"`
}

func (r *Register) normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

type Login struct {
	Email    string `json:"email" validate:"required,request_email,maxbytes=254"`
	Password string `json:"password" validate:"required,maxbytes=72"`
}

func (r *Login) normalize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

type CreateBook struct {
	Name        string `json:"name" validate:"required"`
	Author      string `json:"author"`
	Description string `json:"desc"`
}

func (r *CreateBook) normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Author = strings.TrimSpace(r.Author)
}

// ListBooks uses an inclusive starting ID and a bounded page size.
type ListBooks struct {
	StartID int `query:"start_id" json:"start_id" validate:"gt=0"`
	Limit   int `query:"limit" json:"limit" validate:"gt=0,lte=100"`
}
