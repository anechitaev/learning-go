package main

import "habit-tracker/internal/app"

func main() {
	app := app.NewApplication()

	app.Run()
}
