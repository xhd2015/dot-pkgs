package p

var user *User

type User struct {
	Name string
}

func f() string {
	if user != nil {
		return user.Name
	}
	return ""
}
