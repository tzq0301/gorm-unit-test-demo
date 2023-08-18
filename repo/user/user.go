package user

import "gorm.io/gorm"

type User struct {
	ID   int
	Name string
	Age  int
}

func (User) TableName() string {
	return "user"
}

type Repo struct {
	DB *gorm.DB
}

func (r *Repo) Get(userID int) (User, error) {
	var u User

	err := r.DB.Debug().
		Model(&User{}).
		Where("id = ?", userID).
		First(&u).Error

	if err != nil {
		return User{}, err
	}

	return u, nil
}
