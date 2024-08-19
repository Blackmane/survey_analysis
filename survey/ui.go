package survey

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"
)

func SetupUI(w fyne.Window) {
	label := widget.NewLabel("Upload a CSV file:")
	fileEntry := widget.NewEntry()
	fileEntry.SetPlaceHolder("Path to CSV file")
	fileEntry.Text = "data.csv" // TODO: delete me

	loadButton := widget.NewButton("Load", func() {
		dbPath = fileEntry.Text

		if err := loadDB(); err != nil {
			log.Println("could not check if database exists: ", err)
			return
		}

		showMenu(w)
	})

	createButton := widget.NewButton("Create", func() {
		filePath := fileEntry.Text
		if err := createDB(filePath); err != nil {
			log.Println("could not check if database exists: ", err)
			return
		}

		showMenu(w)
	})

	w.SetContent(container.NewVBox(
		label,
		fileEntry,
		loadButton,
		createButton,
	))
}

func showMenu(w fyne.Window) {
	showDataButton := widget.NewButton("Show Data", func() {
		showData(w)
	})

	showViewsButton := widget.NewButton("Show Views", func() {
		showViews(w)
	})

	createViewButton := widget.NewButton("Create View", func() {
		createViewUI(w)
	})

	backButton := widget.NewButton("Back", func() {
		SetupUI(w)
	})

	w.SetContent(container.NewVBox(
		showDataButton,
		showViewsButton,
		createViewButton,
		backButton,
	))
}

func showData(w fyne.Window) {
	grid := getDataGrid()

	backButton := widget.NewButton("Back", func() {
		showMenu(w)
	})

	w.SetContent(container.NewBorder(nil, backButton, nil, nil, grid))
}

func showViews(w fyne.Window) {
	list := getViews(w)

	backButton := widget.NewButton("Back", func() {
		showMenu(w)
	})

	w.SetContent(container.NewBorder(nil, backButton, nil, nil, list))
}

func createViewUI(w fyne.Window) {
	rows, err := getViewUi()
	if err != nil {
		log.Println("Error querying views:", err)
		dialog.ShowError(err, w)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Println("Error getting columns:", err)
		dialog.ShowError(err, w)
		return
	}
	if len(cols) == 0 {
		log.Println("Dataset without columns:", err)
		dialog.ShowError(err, w)
		return
	}

	checkboxes := make([]*widget.Check, len(cols))
	for i, col := range cols {
		checkboxes[i] = widget.NewCheck(col, nil)
	}

	canvasObjects := make([]fyne.CanvasObject, len(checkboxes))
	for i, check := range checkboxes {
		canvasObjects[i] = check
	}

	viewNameEntry := widget.NewEntry()
	viewNameEntry.SetPlaceHolder("View Name")

	saveButton := widget.NewButton("Save View", func() {
		selectedCols := []string{}
		for i, check := range checkboxes {
			if check.Checked {
				selectedCols = append(selectedCols, cols[i])
			}
		}
		viewName := viewNameEntry.Text
		if viewName == "" {
			dialog.ShowInformation("Error", "View name cannot be empty", w)
			return
		}
		if len(selectedCols) == 0 {
			dialog.ShowInformation("Error", "No columns selected", w)
			return
		}
		selectedColsStr := strings.Join(selectedCols[:], ",")
		err := SaveView(dbPath, viewName, selectedColsStr)
		if err != nil {
			log.Println("Error saving view:", err)
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Success", "View saved successfully", w)
		showMenu(w)
	})

	form := container.NewVBox(
		widget.NewLabel("Select columns to include in view:"),
		container.NewVBox(canvasObjects...),
		widget.NewLabel("View Name:"),
		viewNameEntry,
		saveButton,
	)

	backButton := widget.NewButton("Back", func() {
		showMenu(w)
	})

	w.SetContent(container.NewBorder(nil, backButton, nil, nil, container.NewScroll(form)))
}

func showFilteredData(w fyne.Window, viewName string) {

	rows, err := getViewData(viewName)
	if err != nil {
		log.Println("Error querying filtered data:", err)
		dialog.ShowError(err, w)
		return
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		log.Println("Error getting column names:", err)
		dialog.ShowError(err, w)
		return
	}

	var data [][]string
	data = append(data, colNames)

	for rows.Next() {
		columns := make([]interface{}, len(colNames))
		columnPointers := make([]interface{}, len(colNames))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		err := rows.Scan(columnPointers...)
		if err != nil {
			log.Println("Error scanning row:", err)
			dialog.ShowError(err, w)
			return
		}

		row := make([]string, len(colNames))
		for i, col := range columns {
			if col != nil {
				switch v := col.(type) {
				case int64:
					row[i] = strconv.FormatInt(v, 10)
				case string:
					row[i] = v
				default:
					row[i] = fmt.Sprintf("%v", v)
				}
			} else {
				row[i] = ""
			}
		}

		data = append(data, row)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error with rows:", err)
		dialog.ShowError(err, w)
		return
	}


	gridData := make([]fyne.CanvasObject, len(data)*len(data[0]))
	for i, row := range data {
		for j, cell := range row {
			gridData[i*len(data[0])+j] = widget.NewLabel(cell)
		}
	}

	grid := container.NewGridWithColumns(len(data[0]), gridData...)

	scrollContainer := container.NewHScroll(grid)

	backButton := widget.NewButton("Back", func() {
		showMenu(w)
	})

	w.SetContent(container.NewBorder(nil, backButton, nil, nil, scrollContainer))
}
