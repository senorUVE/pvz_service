package models

import (
	"time"

	"github.com/google/uuid"
)

// type Role string

// const (
// 	RoleModerator Role = "moderator"
// 	RoleEmployee  Role = "employee"
// )

// type City string

// const (
// 	CityMoscow City = "Москва"
// 	CitySPB    City = "Санкт-Петербург"
// 	CityKazan  City = "Казань"
// )

// type Type string

// const (
// 	TypeElectronics Type = "электроника"
// 	TypeClothes     Type = "одежда"
// 	TypeShoes       Type = "обувь"
// )

// type Status string

// const (
// 	StatusInProgress Status = "in_progress"
// 	StatusClose      Status = "close"
// )

type User struct {
	Id       uuid.UUID `json:"id" db:"id"`
	Email    string    `json:"email" db:"email"`
	Password string    `json:"password" db:"password_salt"`
	Role     Role      `json:"role" db:"role"`
}

type PVZ struct {
	Id               uuid.UUID `json:"id" db:"id"`
	RegistrationDate time.Time `json:"registrationDate" db:"registration_date"`
	City             City      `json:"city" db:"city"`
}

type Reception struct {
	Id       uuid.UUID `json:"id" db:"id"`
	DateTime time.Time `json:"dateTime" db:"date_time"`
	PvzId    uuid.UUID `json:"pvzId" db:"pvz_id"`
	Products []Product `json:"products" db:"-"`
	Status   Status    `json:"status" db:"status"`
}

type Product struct {
	Id       uuid.UUID `json:"id" db:"id"`
	DateTime time.Time `json:"dateTime" db:"date_time"`
	Type     Type      `json:"type" db:"type"`
}
