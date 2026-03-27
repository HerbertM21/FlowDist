package main

import (
	"app/db"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
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

type MovimientoCabinet struct {
	ID               int     `json:"id" db:"id"`
	CodigoMovimiento string  `json:"codigo_movimiento" db:"codigo_movimiento"`
	NombreCliente    string  `json:"nombre_cliente" db:"nombre_cliente"`
	Direccion        *string `json:"direccion" db:"direccion"`
	Localidad        *string `json:"localidad" db:"localidad"`
	CantidadCabinets int     `json:"cantidad_cabinets" db:"cantidad_cabinets"`
	Descripcion      *string `json:"descripcion" db:"descripcion"`
	FechaEntrada     *string `json:"fecha_entrada" db:"fecha_entrada"`
	FechaSalida      *string `json:"fecha_salida" db:"fecha_salida"`
	Valor            int     `json:"valor" db:"valor"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
}

type ExportResult struct {
	FilePath string `json:"file_path"`
}

type MovimientoCabinetInput struct {
	NombreCliente    string `json:"nombre_cliente"`
	Direccion        string `json:"direccion"`
	Localidad        string `json:"localidad"`
	CantidadCabinets int    `json:"cantidad_cabinets"`
	Descripcion      string `json:"descripcion"`
	CodigoMovimiento string `json:"codigo_movimiento"`
	FechaEntrada     string `json:"fecha_entrada"`
	FechaSalida      string `json:"fecha_salida"`
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

func (a *App) GetMovimientosCabinets() ([]MovimientoCabinet, error) {
	var items []MovimientoCabinet
	query := `
		SELECT id,
			codigo_movimiento,
			nombre_cliente,
			direccion,
			localidad,
			cantidad_cabinets,
			descripcion,
			TO_CHAR(fecha_entrada, 'YYYY-MM-DD') as fecha_entrada,
			TO_CHAR(fecha_salida, 'YYYY-MM-DD') as fecha_salida,
			COALESCE(valor, 1) as valor,
			TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS') as created_at,
			TO_CHAR(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS') as updated_at
		FROM movimientos_cabinets
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC, id DESC`
	err := a.db.Select(&items, query)
	return items, err
}

func (a *App) CrearMovimientoCabinet(input MovimientoCabinetInput) error {
	codigo := strings.TrimSpace(input.CodigoMovimiento)
	if codigo == "" {
		return fmt.Errorf("el codigo de movimiento es obligatorio")
	}

	nombreCliente := strings.TrimSpace(input.NombreCliente)
	if nombreCliente == "" {
		return fmt.Errorf("el nombre del cliente es obligatorio")
	}

	cantidad := input.CantidadCabinets
	if cantidad <= 0 {
		return fmt.Errorf("la cantidad de cabinets debe ser mayor a 0")
	}

	_, err := a.db.Exec(`
		INSERT INTO movimientos_cabinets (
			nombre_cliente,
			direccion,
			localidad,
			cantidad_cabinets,
			descripcion,
			codigo_movimiento,
			fecha_entrada,
			fecha_salida,
			valor
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1)`,
		nombreCliente,
		nilIfEmpty(input.Direccion),
		nilIfEmpty(input.Localidad),
		cantidad,
		nilIfEmpty(input.Descripcion),
		codigo,
		nilIfEmpty(input.FechaEntrada),
		nilIfEmpty(input.FechaSalida),
	)

	return err
}

func (a *App) UpdateMovimientoCabinet(codigoOriginal string, input MovimientoCabinetInput) error {
	codigoOriginal = strings.TrimSpace(codigoOriginal)
	if codigoOriginal == "" {
		return fmt.Errorf("codigo de movimiento original invalido")
	}

	codigoNuevo := strings.TrimSpace(input.CodigoMovimiento)
	if codigoNuevo == "" {
		return fmt.Errorf("el codigo de movimiento es obligatorio")
	}

	nombreCliente := strings.TrimSpace(input.NombreCliente)
	if nombreCliente == "" {
		return fmt.Errorf("el nombre del cliente es obligatorio")
	}

	cantidad := input.CantidadCabinets
	if cantidad <= 0 {
		return fmt.Errorf("la cantidad de cabinets debe ser mayor a 0")
	}

	_, err := a.db.Exec(`
		UPDATE movimientos_cabinets
		SET
			codigo_movimiento = $1,
			nombre_cliente = $2,
			direccion = $3,
			localidad = $4,
			cantidad_cabinets = $5,
			descripcion = $6,
			fecha_entrada = $7,
			fecha_salida = $8,
			valor = 1,
			updated_at = NOW()
		WHERE codigo_movimiento = $9
			AND deleted_at IS NULL`,
		codigoNuevo,
		nombreCliente,
		nilIfEmpty(input.Direccion),
		nilIfEmpty(input.Localidad),
		cantidad,
		nilIfEmpty(input.Descripcion),
		nilIfEmpty(input.FechaEntrada),
		nilIfEmpty(input.FechaSalida),
		codigoOriginal,
	)

	return err
}

func (a *App) SoftDeleteMovimientoCabinet(codigoMovimiento string) error {
	codigoMovimiento = strings.TrimSpace(codigoMovimiento)
	if codigoMovimiento == "" {
		return fmt.Errorf("codigo de movimiento invalido")
	}

	_, err := a.db.Exec(`
		UPDATE movimientos_cabinets
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE codigo_movimiento = $1
			AND deleted_at IS NULL`, codigoMovimiento)

	return err
}

type exportChequeRow struct {
	NumeroCheque      string  `db:"numero_cheque" json:"numero_cheque"`
	RutCliente        string  `db:"rut_cliente" json:"rut_cliente"`
	NombreCliente     string  `db:"nombre_cliente" json:"nombre_cliente"`
	BancoCheque       string  `db:"banco_cheque" json:"banco_cheque"`
	NumeroFactura     string  `db:"numero_factura" json:"numero_factura"`
	CondicionesPago   string  `db:"condiciones_pago" json:"condiciones_pago"`
	Observaciones     string  `db:"observaciones" json:"observaciones"`
	Monto             float64 `db:"monto" json:"monto"`
	FechaRecepcion    string  `db:"fecha_recepcion" json:"fecha_recepcion"`
	FechaDeposito     string  `db:"fecha_deposito" json:"fecha_deposito"`
	FechaChequeCobrar string  `db:"fecha_cheque_cobrar" json:"fecha_cheque_cobrar"`
	Estado            string  `db:"estado" json:"estado"`
	Vendedor          string  `db:"vendedor" json:"vendedor"`
}

type exportDevueltosRow struct {
	NumeroCheque  string  `db:"numero_cheque" json:"numero_cheque"`
	RutCliente    string  `db:"rut_cliente" json:"rut_cliente"`
	NombreCliente string  `db:"nombre_cliente" json:"nombre_cliente"`
	Plaza         string  `db:"plaza" json:"plaza"`
	Monto         float64 `db:"monto" json:"monto"`
	FechaRegistro string  `db:"fecha_registro" json:"fecha_registro"`
	FechaCheque   string  `db:"fecha_cheque" json:"fecha_cheque"`
	TipoPago      string  `db:"tipo_pago" json:"tipo_pago"`
	Motivo        string  `db:"motivo" json:"motivo"`
	FechaSaldada  string  `db:"fecha_saldada" json:"fecha_saldada"`
	Comentario    string  `db:"comentario" json:"comentario"`
	EstadoPago    string  `db:"estado_pago" json:"estado_pago"`
}

type exportCabinetsRow struct {
	CodigoMovimiento string `db:"codigo_movimiento" json:"codigo_movimiento"`
	NombreCliente    string `db:"nombre_cliente" json:"nombre_cliente"`
	Direccion        string `db:"direccion" json:"direccion"`
	Localidad        string `db:"localidad" json:"localidad"`
	CantidadCabinets int    `db:"cantidad_cabinets" json:"cantidad_cabinets"`
	Descripcion      string `db:"descripcion" json:"descripcion"`
	FechaEntrada     string `db:"fecha_entrada" json:"fecha_entrada"`
	FechaSalida      string `db:"fecha_salida" json:"fecha_salida"`
	Valor            int    `db:"valor" json:"valor"`
}

func (a *App) queryExportChequesRows() ([]exportChequeRow, error) {
	var rows []exportChequeRow
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

func (a *App) queryExportChequesDevueltosRows() ([]exportDevueltosRow, error) {
	var rows []exportDevueltosRow
	query := `
		SELECT
			COALESCE(cd.numero_cheque, '') as numero_cheque,
			cd.rut_cliente,
			COALESCE(cl.razon_social, '') as nombre_cliente,
			COALESCE(cd.plaza, cl.zona, '') as plaza,
			cd.monto,
			COALESCE(TO_CHAR(cd.fecha_registro, 'DD/MM/YYYY'), '') as fecha_registro,
			COALESCE(TO_CHAR(cd.fecha_cheque, 'DD/MM/YYYY'), '') as fecha_cheque,
			COALESCE(cd.tipo_pago, '') as tipo_pago,
			COALESCE(cd.motivo, '') as motivo,
			COALESCE(TO_CHAR(cd.fecha_saldada, 'DD/MM/YYYY'), '') as fecha_saldada,
			COALESCE(cd.comentario, '') as comentario,
			CASE WHEN cd.fecha_saldada IS NOT NULL AND COALESCE(cd.tipo_pago, '') <> '' THEN 'Pagado' ELSE 'Pendiente' END as estado_pago
		FROM cheques_devueltos cd
		LEFT JOIN clientes cl ON cd.rut_cliente = cl.rut
		WHERE cd.deleted_at IS NULL
		ORDER BY cd.created_at DESC, cd.id DESC`
	err := a.db.Select(&rows, query)
	return rows, err
}

func (a *App) queryExportCabinetsRows() ([]exportCabinetsRow, error) {
	var rows []exportCabinetsRow
	query := `
		SELECT
			codigo_movimiento,
			nombre_cliente,
			COALESCE(direccion, '') as direccion,
			COALESCE(localidad, '') as localidad,
			cantidad_cabinets,
			COALESCE(descripcion, '') as descripcion,
			COALESCE(TO_CHAR(fecha_entrada, 'DD/MM/YYYY'), '') as fecha_entrada,
			COALESCE(TO_CHAR(fecha_salida, 'DD/MM/YYYY'), '') as fecha_salida,
			1 as valor
		FROM movimientos_cabinets
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC, codigo_movimiento DESC`
	err := a.db.Select(&rows, query)
	return rows, err
}

func (a *App) ExportChequesExcel() (*ExportResult, error) {
	cheques, err := a.queryExportChequesRows()
	if err != nil {
		return nil, fmt.Errorf("no se pudo consultar cheques para exportar: %w", err)
	}

	devueltos, err := a.queryExportChequesDevueltosRows()
	if err != nil {
		return nil, fmt.Errorf("no se pudo consultar cheques devueltos para exportar: %w", err)
	}

	if a.ctx == nil {
		return nil, fmt.Errorf("contexto de app no disponible")
	}

	filePath, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:           "Exportar Cheques a Excel",
		DefaultFilename: fmt.Sprintf("cheques_%s.xlsx", time.Now().Format("20060102_150405")),
		Filters: []wruntime.FileFilter{
			{DisplayName: "Excel Workbook (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(filePath) == "" {
		return nil, nil
	}

	if err := ensureXLSXExtension(&filePath); err != nil {
		return nil, fmt.Errorf("ruta de exportacion invalida: %w", err)
	}

	f, err := buildChequesWorkbook(cheques, devueltos)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	if err := f.SaveAs(filePath); err != nil {
		return nil, fmt.Errorf("no se pudo guardar el archivo en %q: %w", filePath, err)
	}

	return &ExportResult{FilePath: filePath}, nil
}

func (a *App) ExportCabinetsExcel() (*ExportResult, error) {
	rows, err := a.queryExportCabinetsRows()
	if err != nil {
		return nil, fmt.Errorf("no se pudo consultar control cabinets para exportar: %w", err)
	}

	if a.ctx == nil {
		return nil, fmt.Errorf("contexto de app no disponible")
	}

	filePath, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:           "Exportar Control Cabinets a Excel",
		DefaultFilename: fmt.Sprintf("cabinets_%s.xlsx", time.Now().Format("20060102_150405")),
		Filters: []wruntime.FileFilter{
			{DisplayName: "Excel Workbook (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(filePath) == "" {
		return nil, nil
	}

	if err := ensureXLSXExtension(&filePath); err != nil {
		return nil, fmt.Errorf("ruta de exportacion invalida: %w", err)
	}

	f, err := buildCabinetsWorkbook(rows)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	if err := f.SaveAs(filePath); err != nil {
		return nil, fmt.Errorf("no se pudo guardar el archivo en %q: %w", filePath, err)
	}

	return &ExportResult{FilePath: filePath}, nil
}

func buildChequesWorkbook(cheques []exportChequeRow, devueltos []exportDevueltosRow) (*excelize.File, error) {
	f := excelize.NewFile()

	const sheetCheques = "Cheques"
	defaultSheetName := f.GetSheetName(0)
	if defaultSheetName != "" && defaultSheetName != sheetCheques {
		if err := f.SetSheetName(defaultSheetName, sheetCheques); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("no se pudo renombrar hoja principal: %w", err)
		}
	}
	if defaultSheetName == "" {
		if _, err := f.NewSheet(sheetCheques); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("no se pudo crear hoja %q: %w", sheetCheques, err)
		}
	}

	headersCheques := []string{"N Cheque", "RUT", "Nombre Cliente", "Banco", "N Factura", "Condiciones Pago", "Observaciones", "Monto", "F. Recepcion", "F. Deposito", "F. Cheque Cobrar", "Estado", "Vendedor"}
	rowsCheques := make([][]interface{}, 0, len(cheques))
	for _, r := range cheques {
		rowsCheques = append(rowsCheques, []interface{}{
			r.NumeroCheque,
			r.RutCliente,
			r.NombreCliente,
			r.BancoCheque,
			r.NumeroFactura,
			r.CondicionesPago,
			r.Observaciones,
			r.Monto,
			r.FechaRecepcion,
			r.FechaDeposito,
			r.FechaChequeCobrar,
			r.Estado,
			r.Vendedor,
		})
	}
	if err := writeStyledSheet(f, sheetCheques, headersCheques, rowsCheques); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("no se pudo escribir hoja %q: %w", sheetCheques, err)
	}

	const sheetDevueltos = "Cheques Devueltos"
	if _, err := f.NewSheet(sheetDevueltos); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("no se pudo crear hoja %q: %w", sheetDevueltos, err)
	}

	headersDevueltos := []string{"N Cheque", "RUT", "Nombre Cliente", "Plaza", "Monto", "F. Registro", "F. Cheque", "Tipo Pago", "Motivo", "F. Saldada", "Comentario", "Estado Pago"}
	rowsDevueltos := make([][]interface{}, 0, len(devueltos))
	for _, r := range devueltos {
		rowsDevueltos = append(rowsDevueltos, []interface{}{
			r.NumeroCheque,
			r.RutCliente,
			r.NombreCliente,
			r.Plaza,
			r.Monto,
			r.FechaRegistro,
			r.FechaCheque,
			r.TipoPago,
			r.Motivo,
			r.FechaSaldada,
			r.Comentario,
			r.EstadoPago,
		})
	}
	if err := writeStyledSheet(f, sheetDevueltos, headersDevueltos, rowsDevueltos); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("no se pudo escribir hoja %q: %w", sheetDevueltos, err)
	}

	f.SetActiveSheet(0)
	return f, nil
}

func buildCabinetsWorkbook(rows []exportCabinetsRow) (*excelize.File, error) {
	f := excelize.NewFile()

	const sheetName = "Control Cabinets"
	defaultSheetName := f.GetSheetName(0)
	if defaultSheetName != "" && defaultSheetName != sheetName {
		if err := f.SetSheetName(defaultSheetName, sheetName); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("no se pudo renombrar hoja principal: %w", err)
		}
	}
	if defaultSheetName == "" {
		if _, err := f.NewSheet(sheetName); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("no se pudo crear hoja %q: %w", sheetName, err)
		}
	}

	headers := []string{"Codigo Movimiento", "Nombre Cliente", "Direccion", "Localidad", "Cantidad Cabinets", "Descripcion", "Fecha Entrada", "Fecha Salida", "Valor"}
	data := make([][]interface{}, 0, len(rows))
	for _, r := range rows {
		data = append(data, []interface{}{
			r.CodigoMovimiento,
			r.NombreCliente,
			r.Direccion,
			r.Localidad,
			r.CantidadCabinets,
			r.Descripcion,
			r.FechaEntrada,
			r.FechaSalida,
			r.Valor,
		})
	}
	if err := writeStyledSheet(f, sheetName, headers, data); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("no se pudo escribir hoja %q: %w", sheetName, err)
	}

	f.SetActiveSheet(0)
	return f, nil
}

func writeStyledSheet(f *excelize.File, sheet string, headers []string, rows [][]interface{}) error {
	if len(headers) == 0 {
		return nil
	}

	headerStyleID, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"1F4E78"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	bodyStyleID, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "000000"},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return err
	}

	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, header); err != nil {
			return err
		}
	}

	for r, row := range rows {
		for c, value := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+2)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return err
			}
		}
	}

	lastHeaderCell, err := excelize.CoordinatesToCellName(len(headers), 1)
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", lastHeaderCell, headerStyleID); err != nil {
		return err
	}

	bodyLastRow := len(rows) + 1
	if bodyLastRow > 1 {
		bodyLastCell, err := excelize.CoordinatesToCellName(len(headers), bodyLastRow)
		if err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, "A2", bodyLastCell, bodyStyleID); err != nil {
			return err
		}
	}

	for col := range headers {
		columnName, err := excelize.ColumnNumberToName(col + 1)
		if err != nil {
			return err
		}

		maxLen := len(headers[col])
		for _, row := range rows {
			if col >= len(row) {
				continue
			}
			text := fmt.Sprintf("%v", row[col])
			if len(text) > maxLen {
				maxLen = len(text)
			}
		}

		width := float64(maxLen + 3)
		if width < 14 {
			width = 14
		}
		if width > 56 {
			width = 56
		}
		if err := f.SetColWidth(sheet, columnName, columnName, width); err != nil {
			return err
		}
	}

	if err := f.SetRowHeight(sheet, 1, 24); err != nil {
		return err
	}

	if err := f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{{
			SQRef:      "A2",
			ActiveCell: "A2",
			Pane:       "bottomLeft",
		}},
	}); err != nil {
		return err
	}

	if bodyLastRow >= 1 {
		lastCell, err := excelize.CoordinatesToCellName(len(headers), bodyLastRow)
		if err != nil {
			return err
		}
		rangeRef := "A1:" + lastCell
		if err := f.AutoFilter(sheet, rangeRef, []excelize.AutoFilterOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func ensureXLSXExtension(filePath *string) error {
	if filePath == nil {
		return fmt.Errorf("ruta de archivo invalida")
	}
	path := strings.TrimSpace(*filePath)
	if path == "" {
		return fmt.Errorf("ruta de archivo vacia")
	}

	if !strings.HasSuffix(strings.ToLower(path), ".xlsx") {
		path += ".xlsx"
	}

	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("ruta de archivo invalida")
	}

	parentDir := filepath.Dir(path)
	if parentDir != "" && parentDir != "." {
		if _, err := os.Stat(parentDir); err != nil {
			return fmt.Errorf("directorio destino invalido: %w", err)
		}
	}

	*filePath = path
	return nil
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
