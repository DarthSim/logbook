package main

func main() {
	initLogger()
	defer closeLogger()

	initDB()
	defer closeDB()

	startServer()
}
