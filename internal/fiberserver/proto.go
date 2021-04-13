package fiberserver

import "github.com/go-playground/validator/v10"

type TransID int

type ReqUserInfo struct {
	TID      TransID `json:"tid" validate:"required"`
	Name     string  `json:"name" validate:"required,min=3,max=32"`
	IsActive bool    `json:"isactive" validate:"required,eq=True|eq=False"`
	Email    string  `json:"email" validate:"required,email,min=6,max=32"`
	Job      Job     `json:"job" validate:"dive"`
}

type Job struct {
	Type   string `json:"type" validate:"required,min=3,max=32"`
	Salary int    `json:"salary" validate:"required,number"`
}

type RespUserInfo struct {
	TID       TransID     `json:"tid"`
	Name      string      `json:"name"`
	UserToken interface{} `json:"isactive"`
}

type RespError struct {
	FailedField string
	Tag         string
	Value       string
}

type RespLogin struct {
	Token string `json:"token" validate:"required"`
}

type RespGreet struct {
	//ReqHeader

	Uid   TransID `json:"uid" validate:"required"`
	Title string  `json:"title" validate:"required"`
	Msg   string  `json:"message" validate:"required"`
}

type RespGreetV2 struct {
	UnixNano int64       `json:"utime" validate:"required"`
	Greets   []RespGreet `json:"greets" validate:"dive"`
}

var greets = map[TransID]RespGreet{
	10000001: {
		Uid:   1,
		Title: "Hello Foo",
		Msg:   "Hello To Mr.Foo",
	},
	10000002: {
		Uid:   2,
		Title: "Hello Bar",
		Msg:   "Hello To Ms.Bar",
	},
}

func ValidateStruct(userInfo interface{}) []RespError {
	var errors []RespError
	validate := validator.New()
	err := validate.Struct(userInfo)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element RespError
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, element)
		}
	}
	return errors
}
