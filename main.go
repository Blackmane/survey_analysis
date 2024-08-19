package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"survey-analysis-fyne/survey"
)

func main() {
	a := app.New()
	w := a.NewWindow("Survey Dashboard")

	survey.SetupUI(w)

	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
