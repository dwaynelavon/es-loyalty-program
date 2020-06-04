// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type NewUser struct {
	Username string `json:"username"`
}

type User struct {
	UserID    *string    `json:"userId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	Username  string     `json:"username"`
}

type UserCreateResponse struct {
	UserID   *string `json:"userId"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Status   Status  `json:"status"`
}

type UserDeleteResponse struct {
	UserID *string `json:"userId"`
	Status Status  `json:"status"`
}

type Status string

const (
	StatusAccepted Status = "Accepted"
	StatusRejected Status = "Rejected"
)

var AllStatus = []Status{
	StatusAccepted,
	StatusRejected,
}

func (e Status) IsValid() bool {
	switch e {
	case StatusAccepted, StatusRejected:
		return true
	}
	return false
}

func (e Status) String() string {
	return string(e)
}

func (e *Status) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Status(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Status", str)
	}
	return nil
}

func (e Status) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
