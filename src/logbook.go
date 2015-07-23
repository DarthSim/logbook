package main

func main() {
	prepareConfig()

	initDB()
	defer closeDB()

	startServer()
}
