package models

import "fmt"

type Role string

const (
	RoleModerator Role = "moderator"
	RoleEmployee  Role = "employee"
)

func (Role) Parse(str string) (Role, error) {
	switch str {
	case string(RoleModerator):
		return RoleModerator, nil
	case string(RoleEmployee):
		return RoleEmployee, nil
	}
	return "", fmt.Errorf("invalid role: %s", str)

}

func (r Role) String() string {
	return string(r)
}
