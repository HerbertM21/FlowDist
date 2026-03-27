package main

import (
	"app/db"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse("02/01/2006", dateStr)
	if err == nil {
		return t
	}
	t, err = time.Parse("01-02-06", dateStr)
	if err == nil {
		return t
	}
	return time.Time{}
}

func parseAmount(amountStr string) float64 {
	amountStr = strings.ReplaceAll(amountStr, "$", "")
	amountStr = strings.ReplaceAll(amountStr, ".", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.TrimSpace(amountStr)
	amount, _ := strconv.ParseFloat(amountStr, 64)
	return amount
}

func main() {
	database, err := db.InitDB("")
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	f, err := excelize.OpenFile("../chequescartera.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	sheet := "DETALLE CHEQUES"
	rows, err := f.GetRows(sheet)
	if err != nil {
		log.Fatal(err)
	}

	clientMap := make(map[string]string)
	vendedorMap := make(map[string]int)
	clientCounter := 1

	for i, row := range rows {
		if i < 2 || len(row) < 9 {
			continue
		}

		vendedorNombre := strings.TrimSpace(row[8])
		vendedorID := 0
		if vendedorNombre != "" {
			if id, exists := vendedorMap[vendedorNombre]; exists {
				vendedorID = id
			} else {
				err = database.QueryRow("INSERT INTO vendedores (nombre_completo) VALUES ($1) RETURNING id", vendedorNombre).Scan(&vendedorID)
				if err != nil {
					log.Printf("Error inserting vendedor: %v\n", err)
				} else {
					vendedorMap[vendedorNombre] = vendedorID
				}
			}
		}

		clienteNombre := strings.TrimSpace(row[4])
		clienteRUT := ""
		if clienteNombre != "" {
			if rut, exists := clientMap[clienteNombre]; exists {
				clienteRUT = rut
			} else {
				clienteRUT = fmt.Sprintf("RUT-%04d", clientCounter)
				clientCounter++
				var vID *int
				if vendedorID > 0 {
					vID = &vendedorID
				}
				_, err = database.Exec("INSERT INTO clientes (rut, razon_social, id_vendedor) VALUES ($1, $2, $3)", clienteRUT, clienteNombre, vID)
				if err != nil {
					log.Printf("Error inserting client: %v\n", err)
				} else {
					clientMap[clienteNombre] = clienteRUT
				}
			}
		}

		if clienteRUT == "" {
			continue
		}

		monto := parseAmount(row[5])
		if monto == 0 {
			continue
		}

		numeroCheque := strings.TrimSpace(row[2])
		fechaRecepcion := parseDate(row[1])
		fechaChequeCobrar := parseDate(row[6])

		var fR, fCC *time.Time
		if !fechaRecepcion.IsZero() {
			fR = &fechaRecepcion
		}
		if !fechaChequeCobrar.IsZero() {
			fCC = &fechaChequeCobrar
		}

		var fD *time.Time

		estadoID := 1

		_, err = database.Exec(`
			INSERT INTO cheques (
				numero_cheque, rut_cliente, monto,
				fecha_recepcion, fecha_deposito, fecha_cheque_cobrar,
				id_estado
			) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			numeroCheque, clienteRUT, monto, fR, fD, fCC, estadoID)

		if err != nil {
			log.Printf("Error inserting cheque %s: %v\n", numeroCheque, err)
		}
	}

	log.Println("ETL process completed.")
}
