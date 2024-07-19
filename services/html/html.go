package html

import (
	"bytes"
	"html/template"
	"log"

	"github.com/samber/lo"
)

type Table struct {
	Title   string
	Columns []Column
	Rows    [][]Value
}

type Cell struct {
	Key   string
	Value Value
}

type Value struct {
	Value      interface{}
	Link       string
	IsCheckbox bool
}

type Column struct {
	Key        string
	IsCheckbox bool
}

func PrintTable(title string, table [][]Cell) string {
	tmplTable := mapToTable(title, table)

	t := template.Must(template.New("").Parse(tmpl))

	var body bytes.Buffer
	if err := t.Execute(&body, tmplTable); err != nil {
		log.Fatal(err)
	}
	return body.String()
}

func mapToTable(title string, mapTable [][]Cell) Table {
	if len(mapTable) == 0 {
		return Table{}
	}

	var cols []Column
	var rows [][]Value

	for _, value := range mapTable[0] {
		cols = append(cols, Column{Key: value.Key, IsCheckbox: value.Value.IsCheckbox})
	}

	for _, res := range mapTable {
		vals := make([]Value, len(cols))
		for i, col := range cols {
			colCell, _ := lo.Find(res, func(cell Cell) bool {
				return cell.Key == col.Key
			})
			vals[i] = colCell.Value
		}
		rows = append(rows, vals)
	}
	return Table{Title: title, Rows: rows, Columns: cols}
}
