package common

import "github.com/mahdipeydai/taskmanager-go/constants"

func IsAdmin(roles []string) bool {
	for _, role := range roles {
		if role == constants.AdminRoleName {
			return true
		}
	}

	return false
}
