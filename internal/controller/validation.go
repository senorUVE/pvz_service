package controller

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/senorUVE/pvz_service/internal/dto"
)

func ValidateAuth(request *dto.AuthRequest) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(request.Email) {
		return ErrInvalidEmail
	}

	if len(request.Password) < 8 {
		return ErrShortPassword
	}

	return nil
}

func ValidateCloseLastReceptionRequest(request *dto.CloseLastReceptionRequest) error {
	if request.PvzId == uuid.Nil {
		return ErrInvalidUUID
	}
	return nil
}

func ValidateCloseLastReceptionResponse(response *dto.CloseLastReceptionResponse) error {
	if response.Status != "close" && response.Status != "in_progress" {
		return ErrInvalidStatus
	}

	if response.DateTime.After(time.Now()) {
		return ErrFutureDate
	}

	return nil
}

func ValidatePvzCreateRequest(request *dto.PvzCreateRequest) error {
	validCities := map[string]bool{
		"Москва":          true,
		"Санкт-Петербург": true,
		"Казань":          true,
	}
	if !validCities[request.City] {
		return ErrInvalidCity
	}

	if request.RegistrationDate.After(time.Now()) {
		return ErrFutureDate
	}

	return nil
}

func ValidateAddProductRequest(request *dto.AddProductRequest) error {
	validTypes := map[string]bool{
		"электроника": true,
		"одежда":      true,
		"обувь":       true,
	}
	if !validTypes[request.Type] {
		return ErrInvalidProductType
	}

	if request.PvzId == uuid.Nil {
		return ErrInvalidUUID
	}

	return nil
}

func ValidateGetPvzRequest(request *dto.GetPvzRequest) error {
	if request.Page < 1 {
		return ErrInvalidPage
	}

	if request.Limit < 1 || request.Limit > 30 {
		return ErrInvalidLimit
	}

	if !request.StartDate.IsZero() && !request.EndDate.IsZero() && request.EndDate.Before(request.StartDate) {
		return ErrInvalidDateRange
	}

	currentTime := time.Now().UTC()
	if request.StartDate.After(currentTime) || request.EndDate.After(currentTime) {
		return ErrFutureDate
	}

	return nil
}

func ValidateRegisterRequest(request *dto.RegisterRequest) error {
	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(request.Email) {
		return ErrInvalidEmail
	}

	if len(request.Password) < 8 {
		return ErrWeakPassword
	}

	if request.Role != "moderator" && request.Role != "employee" {
		return ErrInvalidRole
	}

	return nil
}

func ValidateDeleteProductRequest(request *dto.DeleteProductRequest) error {
	if request.PvzId == uuid.Nil || request.PvzId.Version() != 4 {
		return ErrInvalidUUID
	}
	return nil
}

func ValidateReception(reception *dto.ReceptionResponse) error {
	if reception.Status != "in_progress" && reception.Status != "close" {
		return ErrInvalidStatus
	}
	return nil
}

func ValidateProduct(product *dto.ProductResponse) error {
	validTypes := map[string]bool{"электроника": true, "одежда": true, "обувь": true}
	if !validTypes[product.Type] {
		return ErrInvalidProductType
	}
	return nil
}

func ValidatePVZ(pvz *dto.PVZResponse) error {
	validCities := map[string]bool{"Москва": true, "Санкт-Петербург": true, "Казань": true}
	if !validCities[pvz.City] {
		return ErrInvalidCity
	}
	return nil
}
