package main

import (
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestEnsureXLSXExtension(t *testing.T) {
	t.Run("agrega extension faltante", func(t *testing.T) {
		tempDir := t.TempDir()
		p := filepath.Join(tempDir, "reporte_cheques")

		if err := ensureXLSXExtension(&p); err != nil {
			t.Fatalf("ensureXLSXExtension retorno error: %v", err)
		}

		expected := filepath.Join(tempDir, "reporte_cheques.xlsx")
		if p != expected {
			t.Fatalf("ruta inesperada: got=%q want=%q", p, expected)
		}
	})

	t.Run("mantiene extension existente", func(t *testing.T) {
		tempDir := t.TempDir()
		p := filepath.Join(tempDir, "reporte_cheques.xlsx")

		if err := ensureXLSXExtension(&p); err != nil {
			t.Fatalf("ensureXLSXExtension retorno error: %v", err)
		}

		expected := filepath.Join(tempDir, "reporte_cheques.xlsx")
		if p != expected {
			t.Fatalf("ruta inesperada: got=%q want=%q", p, expected)
		}
	})
}

func TestWriteStyledSheet(t *testing.T) {
	t.Run("sin filas de datos", func(t *testing.T) {
		f := excelize.NewFile()
		defer func() {
			_ = f.Close()
		}()

		sheet := "Cheques"
		defaultSheetName := f.GetSheetName(0)
		if defaultSheetName != sheet {
			if defaultSheetName == "" {
				f.NewSheet(sheet)
			} else {
				_ = f.SetSheetName(defaultSheetName, sheet)
			}
		}

		err := writeStyledSheet(f, sheet, []string{"Columna A", "Columna B"}, [][]interface{}{})
		if err != nil {
			t.Fatalf("writeStyledSheet retorno error sin filas: %v", err)
		}

		path := filepath.Join(t.TempDir(), "sin_filas.xlsx")
		if err := f.SaveAs(path); err != nil {
			t.Fatalf("SaveAs retorno error sin filas: %v", err)
		}
	})

	t.Run("con filas de datos", func(t *testing.T) {
		f := excelize.NewFile()
		defer func() {
			_ = f.Close()
		}()

		sheet := "Cheques Devueltos"
		if _, err := f.NewSheet(sheet); err != nil {
			t.Fatalf("NewSheet retorno error: %v", err)
		}

		err := writeStyledSheet(
			f,
			sheet,
			[]string{"Numero", "Cliente", "Monto"},
			[][]interface{}{
				{"001", "Cliente Uno", 1000.0},
				{"002", "Cliente Dos", 2000.0},
			},
		)
		if err != nil {
			t.Fatalf("writeStyledSheet retorno error con filas: %v", err)
		}

		path := filepath.Join(t.TempDir(), "con_filas.xlsx")
		if err := f.SaveAs(path); err != nil {
			t.Fatalf("SaveAs retorno error con filas: %v", err)
		}
	})
}

func TestBuildChequesWorkbook(t *testing.T) {
	cheques := []exportChequeRow{
		{
			NumeroCheque:      "123",
			RutCliente:        "11.111.111-1",
			NombreCliente:     "Cliente Uno",
			BancoCheque:       "Banco Test",
			NumeroFactura:     "F001",
			CondicionesPago:   "30 dias",
			Observaciones:     "Obs",
			Monto:             1000,
			FechaRecepcion:    "01/03/2026",
			FechaDeposito:     "02/03/2026",
			FechaChequeCobrar: "03/03/2026",
			Estado:            "En cartera",
			Vendedor:          "Vendedor Uno",
		},
	}
	devueltos := []exportDevueltosRow{
		{
			NumeroCheque:  "987",
			RutCliente:    "22.222.222-2",
			NombreCliente: "Cliente Dos",
			Plaza:         "Talca",
			Monto:         2000,
			FechaRegistro: "05/03/2026",
			FechaCheque:   "06/03/2026",
			TipoPago:      "Efectivo",
			Motivo:        "Motivo",
			FechaSaldada:  "07/03/2026",
			Comentario:    "Comentario",
			EstadoPago:    "Pagado",
		},
	}

	f, err := buildChequesWorkbook(cheques, devueltos)
	if err != nil {
		t.Fatalf("buildChequesWorkbook retorno error: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if got := f.GetSheetName(0); got != "Cheques" {
		t.Fatalf("hoja activa inesperada: got=%q want=%q", got, "Cheques")
	}

	path := filepath.Join(t.TempDir(), "cheques_workbook.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("SaveAs workbook cheques fallo: %v", err)
	}
}

func TestBuildCabinetsWorkbook(t *testing.T) {
	rows := []exportCabinetsRow{
		{
			CodigoMovimiento: "MOV-001",
			NombreCliente:    "Cliente Uno",
			Direccion:        "Direccion 123",
			Localidad:        "Talca",
			CantidadCabinets: 3,
			Descripcion:      "Movimiento inicial",
			FechaEntrada:     "01/03/2026",
			FechaSalida:      "02/03/2026",
			Valor:            1,
		},
	}

	f, err := buildCabinetsWorkbook(rows)
	if err != nil {
		t.Fatalf("buildCabinetsWorkbook retorno error: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if got := f.GetSheetName(0); got != "Control Cabinets" {
		t.Fatalf("hoja activa inesperada: got=%q want=%q", got, "Control Cabinets")
	}

	path := filepath.Join(t.TempDir(), "cabinets_workbook.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("SaveAs workbook cabinets fallo: %v", err)
	}
}
