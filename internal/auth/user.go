package auth

type User struct {
	Name      string `json:"name"`
	AuthToken string `json:"token"`
}
