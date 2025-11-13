package models

import "errors"

func MapToToken(input map[string]any) (UserToken, error) {
	id, ok := input["ID"].(string)
	if !ok {
		return UserToken{}, errors.New("id error")
	}
	name, ok := input["Name"].(string)
	if !ok {
		return UserToken{}, errors.New("name error")
	}
	ismod, ok := input["IsModer"].(bool)
	if !ok {
		return UserToken{}, errors.New("IsModer error")
	}
	isadm, ok := input["IsAdmin"].(bool)
	if !ok {
		return UserToken{}, errors.New("IsAdmin error")
	}
	iscor, ok := input["IsCore"].(bool)
	if !ok {
		return UserToken{}, errors.New("IsCore error")
	}
	if id == "" || name == "" {
		return UserToken{}, errors.New("id or name is empty")
	}
	return UserToken{
		ID:      id,
		Name:    name,
		IsModer: ismod,
		IsAdmin: isadm,
		IsCore:  iscor,
	}, nil
}
