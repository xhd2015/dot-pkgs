package pkg

func HandleRequest(id int) {
	user := findUser(id)
	log.Println("start")
	saveUser(user)
	validateUser(user)
	log.Println("done")
}
