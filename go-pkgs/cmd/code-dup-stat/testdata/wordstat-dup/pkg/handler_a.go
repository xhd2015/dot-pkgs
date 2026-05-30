package pkg

func ProcessUser(id int) {
	log.Println("start")
	user := findUser(id)
	validateUser(user)
	saveUser(user)
	log.Println("done")
}
