package dto

type BadRequestResponse struct {
	Errors string `json:"errors"`
}

type WrongDataResponse struct {
	Errors string `json:"errors"`
}

type AccessDeniedResponse struct {
	Errors string `json:"errors"`
}

type UnauthorizedResponse struct {
	Errors string `json:"errors"`
}

type InternalServerErrorResponse struct {
	Errors string `json:"errors"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}
