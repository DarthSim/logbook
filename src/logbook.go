package main

func main() {
	prepareConfig()

	initLogger()
	defer closeLogger()

	initDB()
	defer closeDB()

	startServer()
}
