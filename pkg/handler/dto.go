package handler

type userLoginParams struct {
	Email    string `bson:"email" validate:"required"`
	Password string `bson:"password" validate:"required"`
}

type passwordResetParams struct {
	Password        string `bson:"password" validate:"required"`
	ConfirmPassword string `bson:"confirmPassword" validate:"required"`
}
type forgotPasswordParams struct {
	Email string `bson:"email" validate:"required"`
}

type errorResponseParams struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func newErrorResponse(status string, err string) *errorResponseParams {
	return &errorResponseParams{
		Status: status,
		Error:  err,
	}
}
