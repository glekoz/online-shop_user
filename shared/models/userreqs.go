package models

import "github.com/glekoz/online-shop_user/shared/validator"

type RegisterUserReq struct {
	Username string
	Password string
	Email    string
}

func (r *RegisterUserReq) Validate(v *validator.Validator) {
	v.Check(r.Username != "", "username", "must be provided")
	v.Check(len(r.Username) >= 3, "username", "must be at least 3 characters long")
	v.Check(len(r.Username) <= 50, "username", "must not be more than 50 characters long")

	v.Check(r.Password != "", "password", "must be provided")
	v.Check(len(r.Password) >= 6, "password", "must be at least 6 characters long")
	v.Check(len(r.Password) <= 100, "password", "must not be more than 100 characters long")
	v.Check(validator.ValidPassword(r.Password), "password", "must contain at least one uppercase letter, one lowercase letter, one digit, and one special character")

	v.Check(r.Email != "", "email", "must be provided")
	v.Check(len(r.Email) <= 100, "email", "must not be more than 100 characters long")
	v.Check(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")
}

type LoginUserReq struct {
	Email    string
	Password string
}

func (r *LoginUserReq) Validate(v *validator.Validator) {
	v.Check(r.Email != "", "email", "must be provided")
	v.Check(len(r.Email) <= 100, "email", "must not be more than 100 characters long")
	v.Check(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")

	v.Check(r.Password != "", "password", "must be provided")
	v.Check(len(r.Password) >= 6, "password", "must be at least 6 characters long")
	v.Check(len(r.Password) <= 100, "password", "must not be more than 100 characters long")

}
