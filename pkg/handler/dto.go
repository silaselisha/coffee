package handler

type userLoginParams struct {
	Email    string `bson:"email" validate:"required"`
	Password string `bson:"password" validate:"required"`
}

type imageResultParams struct {
	avatarName string
	avatarFile []byte
	err        error
}
