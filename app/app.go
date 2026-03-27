package main

import (
	"app/db"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type App struct {
	ctx          context.Context
	db           *sqlx.DB
	pgListener   *pq.Listener
	listenerStop chan struct{}
	listenerDone chan struct{}
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dsn := db.ResolveDSN("")
	database, err := db.InitDB(dsn)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
	} else {
		a.db = database
	}

	a.startDBSyncListener(dsn)
}

func (a *App) shutdown(ctx context.Context) {
	a.stopDBSyncListener()

	if a.db != nil {
		a.db.Close()
	}
}

type Cliente struct {
	Rut         string  `json:"rut" db:"rut"`
	RazonSocial string  `json:"razon_social" db:"razon_social"`
	IDVendedor  *int    `json:"id_vendedor" db:"id_vendedor"`
	Zona        *string `json:"zona" db:"zona"`
}

type Vendedor struct {
	ID             int     `json:"id" db:"id"`
	NombreCompleto string  `json:"nombre_completo" db:"nombre_completo"`
	Zona           *string `json:"zona" db:"zona"`
}

type EstadoCheque struct {
	ID     int    `json:"id" db:"id"`
	Nombre string `json:"nombre" db:"nombre"`
}

type Cheque struct {
	ID                int     `json:"id" db:"id"`
	NumeroCheque      *string `json:"numero_cheque" db:"numero_cheque"`
	RutCliente        string  `json:"rut_cliente" db:"rut_cliente"`
	NombreCliente     string  `json:"nombre_cliente" db:"nombre_cliente"`
	Monto             float64 `json:"monto" db:"monto"`
	FechaRecepcion    *string `json:"fecha_recepcion" db:"fecha_recepcion"`
	FechaDeposito     *string `json:"fecha_deposito" db:"fecha_deposito"`
	FechaChequeCobrar *string `json:"fecha_cheque_cobrar" db:"fecha_cheque_cobrar"`
	BancoCheque       *string `json:"banco_cheque" db:"banco_cheque"`
	NumeroFactura     *string `json:"numero_factura" db:"numero_factura"`
	CondicionesPago   *string `json:"condiciones_pago" db:"condiciones_pago"`
	Observaciones     *string `json:"observaciones" db:"observaciones"`
	IDEstado          *int    `json:"id_estado" db:"id_estado"`
	EstadoNombre      *string `json:"estado_nombre" db:"estado_nombre"`
	NombreVendedor    *string `json:"nombre_vendedor" db:"nombre_vendedor"`
}

type ChequeInput struct {
	NumeroCheque      string  `json:"numero_cheque"`
	RutCliente        string  `json:"rut_cliente"`
	NombreCliente     string  `json:"nombre_cliente"`
	Monto             float64 `json:"monto"`
	FechaRecepcion    string  `json:"fecha_recepcion"`
	FechaDeposito     string  `json:"fecha_deposito"`
	FechaChequeCobrar string  `json:"fecha_cheque_cobrar"`
	BancoCheque       string  `json:"banco_cheque"`
	NumeroFactura     string  `json:"numero_factura"`
	CondicionesPago   string  `json:"condiciones_pago"`
	Observaciones     string  `json:"observaciones"`
	IDEstado          int     `json:"id_estado"`
}

type ClienteMoroso struct {
	Rut             string  `json:"rut" db:"rut"`
	RazonSocial     string  `json:"razon_social" db:"razon_social"`
	NombreVendedor  *string `json:"nombre_vendedor" db:"nombre_vendedor"`
	TotalDeuda      float64 `json:"total_deuda" db:"total_deuda"`
	ChequesVencidos int     `json:"cheques_vencidos" db:"cheques_vencidos"`
}

type ChequeDevuelto struct {
	ID            int     `json:"id" db:"id"`
	IDCheque      *int    `json:"id_cheque" db:"id_cheque"`
	NumeroCheque  *string `json:"numero_cheque" db:"numero_cheque"`
	NumeroFactura *string `json:"numero_factura" db:"numero_factura"`
	Monto         float64 `json:"monto" db:"monto"`
	RutCliente    string  `json:"rut_cliente" db:"rut_cliente"`
	NombreCliente string  `json:"nombre_cliente" db:"nombre_cliente"`
	Plaza         *string `json:"plaza" db:"plaza"`
	FechaRegistro *string `json:"fecha_registro" db:"fecha_registro"`
	FechaCheque   *string `json:"fecha_cheque" db:"fecha_cheque"`
	TipoPago      *string `json:"tipo_pago" db:"tipo_pago"`
	Motivo        *string `json:"motivo" db:"motivo"`
	FechaSaldada  *string `json:"fecha_saldada" db:"fecha_saldada"`
	Comentario    *string `json:"comentario" db:"comentario"`
	EstadoPago    string  `json:"estado_pago" db:"estado_pago"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
}

type ChequeDevueltoInput struct {
	IDCheque      *int    `json:"id_cheque"`
	RutCliente    string  `json:"rut_cliente"`
	NombreCliente string  `json:"nombre_cliente"`
	NumeroCheque  string  `json:"numero_cheque"`
	NumeroFactura string  `json:"numero_factura"`
	Monto         float64 `json:"monto"`
	Plaza         string  `json:"plaza"`
	FechaRegistro string  `json:"fecha_registro"`
	FechaCheque   string  `json:"fecha_cheque"`
	TipoPago      string  `json:"tipo_pago"`
	Motivo        string  `json:"motivo"`
	FechaSaldada  string  `json:"fecha_saldada"`
	Comentario    string  `json:"comentario"`
}

func (a *App) GetClientes() ([]Cliente, error) {
	var clientes []Cliente
	err := a.db.Select(&clientes, "SELECT rut, razon_social, id_vendedor, zona FROM clientes WHERE deleted_at IS NULL ORDER BY razon_social")
	return clientes, err
}

func (a *App) GetClienteByRut(rut string) (*Cliente, error) {
	var c Cliente
	err := a.db.Get(&c, "SELECT rut, razon_social, id_vendedor, zona FROM clientes WHERE rut = $1 AND deleted_at IS NULL", rut)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (a *App) GetVendedores() ([]Vendedor, error) {
	var vendedores []Vendedor
	err := a.db.Select(&vendedores, "SELECT id, nombre_completo, zona FROM vendedores ORDER BY nombre_completo")
	return vendedores, err
}

func (a *App) GetEstados() ([]EstadoCheque, error) {
	var estados []EstadoCheque
	err := a.db.Select(&estados, "SELECT id, nombre FROM estados_cheque ORDER BY id")
	return estados, err
}

func (a *App) GetCheques() ([]Cheque, error) {
	var cheques []Cheque
	query := `
		SELECT c.id, c.numero_cheque, c.rut_cliente, cl.razon_social as nombre_cliente, c.monto,
			TO_CHAR(c.fecha_recepcion, 'YYYY-MM-DD') as fecha_recepcion,
			TO_CHAR(c.fecha_deposito, 'YYYY-MM-DD') as fecha_deposito,
			TO_CHAR(c.fecha_cheque_cobrar, 'YYYY-MM-DD') as fecha_cheque_cobrar,
			c.banco_cheque, c.numero_factura, c.condiciones_pago, c.observaciones,
			c.id_estado, e.nombre as estado_nombre, v.nombre_completo as nombre_vendedor
		FROM cheques c
		JOIN clientes cl ON c.rut_cliente = cl.rut
		LEFT JOIN estados_cheque e ON c.id_estado = e.id
		LEFT JOIN vendedores v ON cl.id_vendedor = v.id
		WHERE c.deleted_at IS NULL
		ORDER BY c.fecha_cheque_cobrar DESC NULLS LAST, c.id DESC`
	err := a.db.Select(&cheques, query)
	return cheques, err
}

func (a *App) CrearCheque(input ChequeInput) error {
	rutCliente, err := a.resolveOrCreateCliente(input.RutCliente, input.NombreCliente)
	if err != nil {
		return err
	}

	_, err = a.db.Exec(`
		INSERT INTO cheques (
			numero_cheque, rut_cliente, monto,
			fecha_recepcion, fecha_deposito, fecha_cheque_cobrar,
			banco_cheque, numero_factura, condiciones_pago, observaciones,
			id_estado
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		nilIfEmpty(input.NumeroCheque), rutCliente, input.Monto,
		nilIfEmpty(input.FechaRecepcion), nilIfEmpty(input.FechaDeposito), nilIfEmpty(input.FechaChequeCobrar),
		nilIfEmpty(input.BancoCheque), nilIfEmpty(input.NumeroFactura), nilIfEmpty(input.CondicionesPago), nilIfEmpty(input.Observaciones),
		input.IDEstado)
	return err
}

func (a *App) UpdateCheque(id int, input ChequeInput) error {
	if id <= 0 {
		return fmt.Errorf("id de cheque invalido")
	}

	var currentEstado sql.NullInt64
	err := a.db.Get(&currentEstado, `
		SELECT id_estado
		FROM cheques
		WHERE id = $1
			AND deleted_at IS NULL`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("cheque no encontrado")
		}
		return err
	}

	if currentEstado.Valid && currentEstado.Int64 == 3 {
		return fmt.Errorf("el cheque protestado/devuelto solo puede gestionarse en la seccion Cheques devueltos")
	}

	rutCliente, err := a.resolveOrCreateCliente(input.RutCliente, input.NombreCliente)
	if err != nil {
		return err
	}

	_, err = a.db.Exec(`
		UPDATE cheques
		SET
			numero_cheque = $1,
			rut_cliente = $2,
			monto = $3,
			fecha_recepcion = $4,
			fecha_deposito = $5,
			fecha_cheque_cobrar = $6,
			banco_cheque = $7,
			numero_factura = $8,
			condiciones_pago = $9,
			observaciones = $10,
			id_estado = $11,
			updated_at = NOW()
		WHERE id = $12
			AND deleted_at IS NULL`,
		nilIfEmpty(input.NumeroCheque),
		rutCliente,
		input.Monto,
		nilIfEmpty(input.FechaRecepcion),
		nilIfEmpty(input.FechaDeposito),
		nilIfEmpty(input.FechaChequeCobrar),
		nilIfEmpty(input.BancoCheque),
		nilIfEmpty(input.NumeroFactura),
		nilIfEmpty(input.CondicionesPago),
		nilIfEmpty(input.Observaciones),
		input.IDEstado,
		id,
	)
	return err
}

func (a *App) UpdateEstadoCheque(id int, nuevoEstadoID int) error {
	if id <= 0 {
		return fmt.Errorf("id de cheque invalido")
	}

	var currentEstado sql.NullInt64
	err := a.db.Get(&currentEstado, `
		SELECT id_estado
		FROM cheques
		WHERE id = $1
			AND deleted_at IS NULL`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("cheque no encontrado")
		}
		return err
	}

	if currentEstado.Valid && currentEstado.Int64 == 3 {
		return fmt.Errorf("el cheque protestado/devuelto solo puede gestionarse en la seccion Cheques devueltos")
	}

	_, err = a.db.Exec("UPDATE cheques SET id_estado = $1, updated_at = NOW() WHERE id = $2", nuevoEstadoID, id)
	return err
}

func (a *App) SoftDeleteCheque(id int) error {
	_, err := a.db.Exec("UPDATE cheques SET deleted_at = NOW() WHERE id = $1", id)
	return err
}

func (a *App) GetMorosos() ([]ClienteMoroso, error) {
	var morosos []ClienteMoroso
	query := `
		SELECT cl.rut, cl.razon_social, v.nombre_completo as nombre_vendedor,
			SUM(c.monto) as total_deuda, COUNT(c.id) as cheques_vencidos
		FROM cheques c
		JOIN clientes cl ON c.rut_cliente = cl.rut
		LEFT JOIN vendedores v ON cl.id_vendedor = v.id
		WHERE c.deleted_at IS NULL
			AND c.fecha_cheque_cobrar < CURRENT_DATE
			AND (c.id_estado = 1 OR c.id_estado IS NULL)
		GROUP BY cl.rut, cl.razon_social, v.nombre_completo
		ORDER BY total_deuda DESC`
	err := a.db.Select(&morosos, query)
	return morosos, err
}

func (a *App) GetChequesDevueltos() ([]ChequeDevuelto, error) {
	var items []ChequeDevuelto
	query := `
		SELECT cd.id, cd.id_cheque, cd.numero_cheque, cd.numero_factura, cd.monto,
			cd.rut_cliente, cl.razon_social as nombre_cliente,
			COALESCE(cd.plaza, cl.zona) as plaza,
			TO_CHAR(cd.fecha_registro, 'YYYY-MM-DD') as fecha_registro,
			TO_CHAR(cd.fecha_cheque, 'YYYY-MM-DD') as fecha_cheque,
			cd.tipo_pago, cd.motivo,
			TO_CHAR(cd.fecha_saldada, 'YYYY-MM-DD') as fecha_saldada,
			cd.comentario,
			CASE WHEN cd.fecha_saldada IS NOT NULL AND COALESCE(cd.tipo_pago, '') <> '' THEN 'Pagado' ELSE 'Pendiente' END as estado_pago,
			TO_CHAR(cd.created_at, 'YYYY-MM-DD"T"HH24:MI:SS') as created_at,
			TO_CHAR(cd.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS') as updated_at
		FROM cheques_devueltos cd
		JOIN clientes cl ON cd.rut_cliente = cl.rut
		WHERE cd.deleted_at IS NULL
		ORDER BY cd.created_at DESC, cd.id DESC`
	err := a.db.Select(&items, query)
	return items, err
}

func (a *App) CrearChequeDevuelto(input ChequeDevueltoInput) error {
	rut, err := a.resolveOrCreateCliente(input.RutCliente, input.NombreCliente)
	if err != nil {
		return err
	}
	if input.Monto <= 0 {
		return fmt.Errorf("el monto debe ser mayor a 0")
	}
	if err := validateChequeDevueltoPagoData(input.TipoPago, input.FechaSaldada); err != nil {
		return err
	}

	tipoPago, err := normalizeTipoPago(input.TipoPago)
	if err != nil {
		return err
	}

	_, err = a.db.Exec(`
		INSERT INTO cheques_devueltos (
			id_cheque, rut_cliente, numero_cheque, numero_factura,
			monto, plaza, fecha_registro, fecha_cheque, tipo_pago, motivo, fecha_saldada, comentario
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		input.IDCheque,
		rut,
		nilIfEmpty(input.NumeroCheque),
		nilIfEmpty(input.NumeroFactura),
		input.Monto,
		nilIfEmpty(input.Plaza),
		nilIfEmpty(input.FechaRegistro),
		nilIfEmpty(input.FechaCheque),
		tipoPago,
		nilIfEmpty(input.Motivo),
		nilIfEmpty(input.FechaSaldada),
		nilIfEmpty(input.Comentario),
	)
	return err
}

func (a *App) UpdateChequeDevuelto(id int, input ChequeDevueltoInput) error {
	if id <= 0 {
		return fmt.Errorf("id de cheque devuelto invalido")
	}
	if input.Monto <= 0 {
		return fmt.Errorf("el monto debe ser mayor a 0")
	}

	rut, err := a.resolveOrCreateCliente(input.RutCliente, input.NombreCliente)
	if err != nil {
		return err
	}
	if err := validateChequeDevueltoPagoData(input.TipoPago, input.FechaSaldada); err != nil {
		return err
	}

	tipoPago, err := normalizeTipoPago(input.TipoPago)
	if err != nil {
		return err
	}

	_, err = a.db.Exec(`
		UPDATE cheques_devueltos
		SET
			rut_cliente = $1,
			numero_cheque = $2,
			monto = $3,
			plaza = $4,
			fecha_registro = $5,
			fecha_cheque = $6,
			tipo_pago = $7,
			motivo = $8,
			fecha_saldada = $9,
			comentario = $10,
			updated_at = NOW()
		WHERE id = $11
			AND deleted_at IS NULL`,
		rut,
		nilIfEmpty(input.NumeroCheque),
		input.Monto,
		nilIfEmpty(input.Plaza),
		nilIfEmpty(input.FechaRegistro),
		nilIfEmpty(input.FechaCheque),
		tipoPago,
		nilIfEmpty(input.Motivo),
		nilIfEmpty(input.FechaSaldada),
		nilIfEmpty(input.Comentario),
		id,
	)
	return err
}

func (a *App) SoftDeleteChequeDevuelto(id int) error {
	if id <= 0 {
		return fmt.Errorf("id de cheque devuelto invalido")
	}

	_, err := a.db.Exec(`
		UPDATE cheques_devueltos
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

type ExportRow struct {
	NumeroCheque      string  `json:"numero_cheque"`
	RutCliente        string  `json:"rut_cliente"`
	NombreCliente     string  `json:"nombre_cliente"`
	BancoCheque       string  `json:"banco_cheque"`
	NumeroFactura     string  `json:"numero_factura"`
	CondicionesPago   string  `json:"condiciones_pago"`
	Observaciones     string  `json:"observaciones"`
	Monto             float64 `json:"monto"`
	FechaRecepcion    string  `json:"fecha_recepcion"`
	FechaDeposito     string  `json:"fecha_deposito"`
	FechaChequeCobrar string  `json:"fecha_cheque_cobrar"`
	Estado            string  `json:"estado"`
	Vendedor          string  `json:"vendedor"`
}

func (a *App) ExportCheques() ([]ExportRow, error) {
	var rows []ExportRow
	query := `
		SELECT COALESCE(c.numero_cheque, '') as numero_cheque, c.rut_cliente,
			cl.razon_social as nombre_cliente,
			COALESCE(c.banco_cheque, '') as banco_cheque,
			COALESCE(c.numero_factura, '') as numero_factura,
			COALESCE(c.condiciones_pago, '') as condiciones_pago,
			COALESCE(c.observaciones, '') as observaciones,
			c.monto,
			COALESCE(TO_CHAR(c.fecha_recepcion, 'DD/MM/YYYY'), '') as fecha_recepcion,
			COALESCE(TO_CHAR(c.fecha_deposito, 'DD/MM/YYYY'), '') as fecha_deposito,
			COALESCE(TO_CHAR(c.fecha_cheque_cobrar, 'DD/MM/YYYY'), '') as fecha_cheque_cobrar,
			COALESCE(e.nombre, '') as estado,
			COALESCE(v.nombre_completo, '') as vendedor
		FROM cheques c
		JOIN clientes cl ON c.rut_cliente = cl.rut
		LEFT JOIN estados_cheque e ON c.id_estado = e.id
		LEFT JOIN vendedores v ON cl.id_vendedor = v.id
		WHERE c.deleted_at IS NULL
		ORDER BY c.fecha_cheque_cobrar DESC NULLS LAST, c.id DESC`
	err := a.db.Select(&rows, query)
	return rows, err
}

func (a *App) CrearCliente(rut, razonSocial string, idVendedor int, zona string) error {
	normalizedRUT, err := normalizeAndValidateRUT(rut)
	if err != nil {
		return err
	}

	var vID *int
	if idVendedor > 0 {
		vID = &idVendedor
	}
	var z *string
	if zona != "" {
		z = &zona
	}
	_, err = a.db.Exec("INSERT INTO clientes (rut, razon_social, id_vendedor, zona) VALUES ($1, $2, $3, $4)", normalizedRUT, razonSocial, vID, z)
	return err
}

func (a *App) GetServerTime() string {
	return time.Now().Format("2006-01-02")
}

func (a *App) resolveOrCreateCliente(raw, nombreCliente string) (string, error) {
	rutInput, err := normalizeAndValidateRUT(raw)
	if err != nil {
		return "", err
	}
	nombre := strings.TrimSpace(nombreCliente)
	if nombre == "" {
		nombre = fmt.Sprintf("Cliente %s", rutInput)
	}

	var rutCliente string
	err = a.db.Get(&rutCliente, `
		SELECT rut
		FROM clientes
		WHERE deleted_at IS NULL
			AND upper(rut) = upper($1)
		LIMIT 1`, rutInput)
	if err == nil {
		_, _ = a.db.Exec(`
			UPDATE clientes
			SET razon_social = CASE
				WHEN (razon_social IS NULL OR razon_social = '' OR razon_social LIKE 'Cliente %') AND $2 <> '' THEN $2
				ELSE razon_social
			END,
			updated_at = NOW()
			WHERE rut = $1`, rutCliente, nombre)
		return rutCliente, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	normalizedInput := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(rutInput, ".", ""), "-", ""))
	err = a.db.Get(&rutCliente, `
		SELECT rut
		FROM clientes
		WHERE deleted_at IS NULL
			AND upper(regexp_replace(rut, '[^0-9A-Z]', '', 'g')) = $1
		LIMIT 1`, normalizedInput)
	if err == nil {
		_, _ = a.db.Exec(`
			UPDATE clientes
			SET razon_social = CASE
				WHEN (razon_social IS NULL OR razon_social = '' OR razon_social LIKE 'Cliente %') AND $2 <> '' THEN $2
				ELSE razon_social
			END,
			updated_at = NOW()
			WHERE rut = $1`, rutCliente, nombre)
		return rutCliente, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	_, err = a.db.Exec(`
		INSERT INTO clientes (rut, razon_social, created_at, updated_at, deleted_at)
		VALUES ($1, $2, NOW(), NOW(), NULL)
		ON CONFLICT (rut) DO UPDATE
		SET deleted_at = NULL,
			razon_social = CASE
				WHEN clientes.razon_social IS NULL OR clientes.razon_social = '' OR clientes.razon_social LIKE 'Cliente %' THEN EXCLUDED.razon_social
				ELSE clientes.razon_social
			END,
			updated_at = NOW()`, rutInput, nombre)
	if err != nil {
		return "", err
	}

	return rutInput, nil
}

func (a *App) ResolveOrCreateCliente(rut, nombreCliente string) (string, error) {
	return a.resolveOrCreateCliente(rut, nombreCliente)
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func normalizeAndValidateRUT(raw string) (string, error) {
	rut := strings.ToUpper(strings.TrimSpace(raw))
	rut = strings.ReplaceAll(rut, ".", "")

	if rut == "" {
		return "", fmt.Errorf("el RUT del cliente es obligatorio")
	}

	if strings.HasPrefix(rut, "RUT-") {
		n := strings.TrimPrefix(rut, "RUT-")
		if n == "" {
			return "", fmt.Errorf("formato de RUT invalido. Usa 21189898-4")
		}
		for _, ch := range n {
			if ch < '0' || ch > '9' {
				return "", fmt.Errorf("formato de RUT invalido. Usa 21189898-4")
			}
		}
		return "RUT-" + n, nil
	}

	parts := strings.Split(rut, "-")
	if len(parts) != 2 {
		return "", fmt.Errorf("formato de RUT invalido. Usa 21189898-4")
	}

	body := parts[0]
	dv := parts[1]

	if len(body) < 7 || len(body) > 8 || len(dv) != 1 {
		return "", fmt.Errorf("formato de RUT invalido. Usa 21189898-4")
	}

	for _, ch := range body {
		if ch < '0' || ch > '9' {
			return "", fmt.Errorf("formato de RUT invalido. Usa 21189898-4")
		}
	}

	if !((dv[0] >= '0' && dv[0] <= '9') || dv == "K") {
		return "", fmt.Errorf("formato de RUT invalido. Usa 21189898-4")
	}

	return body + "-" + dv, nil
}

func normalizeTipoPago(raw string) (interface{}, error) {
	tipo := strings.TrimSpace(strings.ToLower(raw))
	switch tipo {
	case "", "pendiente":
		return nil, nil
	case "transferencia", "efectivo", "cheque":
		return strings.Title(tipo), nil
	default:
		return nil, fmt.Errorf("tipo de pago invalido. Debe ser Transferencia, Efectivo o Cheque")
	}
}

func validateChequeDevueltoPagoData(tipoPagoRaw, fechaSaldadaRaw string) error {
	tipoPago := strings.TrimSpace(strings.ToLower(tipoPagoRaw))
	fechaSaldada := strings.TrimSpace(fechaSaldadaRaw)

	if (tipoPago == "" || tipoPago == "pendiente") && fechaSaldada != "" {
		return fmt.Errorf("si informas fecha saldada debes seleccionar tipo de pago")
	}
	if tipoPago != "" && tipoPago != "pendiente" && fechaSaldada == "" {
		return fmt.Errorf("si seleccionas tipo de pago debes informar fecha saldada")
	}

	return nil
}
