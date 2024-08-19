package survey

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)


func getDataGrid() *fyne.Container {
	rows, err := getData()
	if err != nil {
		log.Println("Error getting data:", err)
		return nil
	}

	cols, err := rows.Columns()
	if err != nil {
		log.Println("Error getting columns:", err)
		return nil
	}

	var data [][]string
	data = append(data, cols)

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		err := rows.Scan(columnPointers...)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil
		}

		row := make([]string, len(cols))
		for i, col := range columns {
			if i > 0 {
				if col != nil {
					row[i-1] = col.(string)
				} else {
					row[i-1] = ""
				}
			}
		}

		data = append(data, row)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error with rows:", err)
		return nil
	}

	gridData := make([]fyne.CanvasObject, len(data)*len(data[0]))
	for i, row := range data {
		for j, cell := range row {
			gridData[i*len(data[0])+j] = widget.NewLabel(cell)
		}
	}

	grid := container.NewGridWithColumns(len(data[0]), gridData...)
	return grid
}

func getViews(w fyne.Window) *widget.List {
	rows, err := getViewList()
	if err != nil {
		log.Println("Error querying views:", err)
		dialog.ShowError(err, w)
		// TODO: spostare showerror
		return nil
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var viewName string
		err := rows.Scan(&viewName)
		if err != nil {
			log.Println("Error scanning row:", err)
			dialog.ShowError(err, w)
			return nil
		}
		views = append(views, viewName)
	}

	list := widget.NewList(
		func() int {
			return len(views)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(views[i])
		})

	list.OnSelected = func(id widget.ListItemID) {
		viewName := views[id]
		showFilteredData(w, viewName)
	}
	return list
}
