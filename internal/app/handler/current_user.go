package handler

import (
	"sync"

	"awesomeProject/internal/app/ds"
)

// creatorUserID — текущий пользователь-создатель зафиксирован константой до появления авторизации (ЛР4)
const creatorUserID uint = 1

var (
	currentUser   *ds.User
	currentUserMu sync.Mutex
)

// CurrentUser — функция-singleton: пользователь загружается из БД один раз
// при первом успешном обращении, дальше все методы получают тот же объект
func (h *Handler) CurrentUser() (ds.User, error) {
	currentUserMu.Lock()
	defer currentUserMu.Unlock()

	if currentUser == nil {
		user, err := h.Repository.GetUserByID(creatorUserID)
		if err != nil {
			return ds.User{}, err
		}
		currentUser = &user
	}
	return *currentUser, nil
}
