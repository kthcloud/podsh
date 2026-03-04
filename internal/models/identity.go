package models

type Identity struct {
	UserID   string `bson:"id" json:"userID"`
	Username string `bson:"username" json:"username"`
	// Role     string `bson:"role" json:"role"`
	Admin bool `bson:"isAdmin" json:"admin"`
}
