-- Schema for MVP 1: Módulo de Cobranzas y Cheques

CREATE TABLE IF NOT EXISTS estados_cheque (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(50) NOT NULL UNIQUE
);

INSERT INTO estados_cheque (nombre) VALUES 
('En cartera'), 
('Depositado'), 
('Protestado/Devuelto')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS vendedores (
    id SERIAL PRIMARY KEY,
    nombre_completo VARCHAR(255) NOT NULL,
    zona VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS clientes (
    rut VARCHAR(15) PRIMARY KEY,
    razon_social VARCHAR(255) NOT NULL,
    id_vendedor INTEGER REFERENCES vendedores(id),
    zona VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS cheques (
    id SERIAL PRIMARY KEY,
    numero_cheque VARCHAR(50),
    rut_cliente VARCHAR(15) REFERENCES clientes(rut),
    monto NUMERIC(12, 2) NOT NULL,
    fecha_recepcion DATE,
    fecha_deposito DATE,
    fecha_cheque_cobrar DATE,
    banco_cheque VARCHAR(120),
    numero_factura VARCHAR(80),
    condiciones_pago VARCHAR(255),
    observaciones TEXT,
    id_estado INTEGER REFERENCES estados_cheque(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE cheques ADD COLUMN IF NOT EXISTS fecha_recepcion DATE;
ALTER TABLE cheques ADD COLUMN IF NOT EXISTS fecha_deposito DATE;
ALTER TABLE cheques ADD COLUMN IF NOT EXISTS fecha_cheque_cobrar DATE;
ALTER TABLE cheques ADD COLUMN IF NOT EXISTS banco_cheque VARCHAR(120);
ALTER TABLE cheques ADD COLUMN IF NOT EXISTS numero_factura VARCHAR(80);
ALTER TABLE cheques ADD COLUMN IF NOT EXISTS condiciones_pago VARCHAR(255);
ALTER TABLE cheques ADD COLUMN IF NOT EXISTS observaciones TEXT;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'cheques'
          AND column_name = 'fecha_emision'
    ) THEN
        UPDATE cheques SET fecha_recepcion = COALESCE(fecha_recepcion, fecha_emision);
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'cheques'
          AND column_name = 'fecha_vencimiento'
    ) THEN
        UPDATE cheques SET fecha_cheque_cobrar = COALESCE(fecha_cheque_cobrar, fecha_vencimiento);
    END IF;
END $$;

ALTER TABLE cheques DROP COLUMN IF EXISTS fecha_emision;
ALTER TABLE cheques DROP COLUMN IF EXISTS fecha_vencimiento;
ALTER TABLE cheques DROP COLUMN IF EXISTS tipo_pago;

CREATE TABLE IF NOT EXISTS cheques_devueltos (
    id SERIAL PRIMARY KEY,
    id_cheque INTEGER UNIQUE REFERENCES cheques(id) ON DELETE SET NULL,
    rut_cliente VARCHAR(15) NOT NULL REFERENCES clientes(rut),
    numero_cheque VARCHAR(50),
    numero_factura VARCHAR(80),
    monto NUMERIC(12, 2) NOT NULL,
    plaza VARCHAR(120),
    fecha_registro DATE DEFAULT CURRENT_DATE,
    fecha_cheque DATE,
    tipo_pago VARCHAR(20),
    motivo TEXT,
    fecha_saldada DATE,
    comentario TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS movimientos_cabinets (
    id SERIAL,
    nombre_cliente VARCHAR(255) NOT NULL,
    direccion VARCHAR(255),
    localidad VARCHAR(120),
    cantidad_cabinets INTEGER NOT NULL DEFAULT 1,
    descripcion TEXT,
    codigo_movimiento VARCHAR(120) PRIMARY KEY,
    fecha_entrada DATE,
    fecha_salida DATE,
    valor INTEGER NOT NULL DEFAULT 1 CHECK (valor = 1),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS nombre_cliente VARCHAR(255);
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS direccion VARCHAR(255);
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS localidad VARCHAR(120);
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS cantidad_cabinets INTEGER DEFAULT 1;
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS descripcion TEXT;
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS codigo_movimiento VARCHAR(120);
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS fecha_entrada DATE;
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS fecha_salida DATE;
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS valor INTEGER DEFAULT 1;
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE movimientos_cabinets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'movimientos_cabinets'
          AND column_name = 'id'
    ) THEN
        IF EXISTS (
            SELECT 1
            FROM information_schema.table_constraints tc
            WHERE tc.table_schema = 'public'
              AND tc.table_name = 'movimientos_cabinets'
              AND tc.constraint_type = 'PRIMARY KEY'
              AND tc.constraint_name = 'movimientos_cabinets_pkey'
        ) THEN
            ALTER TABLE movimientos_cabinets DROP CONSTRAINT movimientos_cabinets_pkey;
        END IF;

        ALTER TABLE movimientos_cabinets
            ADD CONSTRAINT movimientos_cabinets_pkey PRIMARY KEY (codigo_movimiento);
    END IF;
END $$;

ALTER TABLE movimientos_cabinets
    ALTER COLUMN nombre_cliente SET NOT NULL,
    ALTER COLUMN codigo_movimiento SET NOT NULL,
    ALTER COLUMN cantidad_cabinets SET NOT NULL,
    ALTER COLUMN cantidad_cabinets SET DEFAULT 1,
    ALTER COLUMN valor SET NOT NULL,
    ALTER COLUMN valor SET DEFAULT 1;

UPDATE movimientos_cabinets
SET valor = 1
WHERE valor IS DISTINCT FROM 1;

ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS id_cheque INTEGER;
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS rut_cliente VARCHAR(15);
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS numero_cheque VARCHAR(50);
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS numero_factura VARCHAR(80);
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS monto NUMERIC(12, 2);
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS plaza VARCHAR(120);
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS fecha_registro DATE;
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS fecha_cheque DATE;
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS tipo_pago VARCHAR(20);
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS motivo TEXT;
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS fecha_saldada DATE;
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS comentario TEXT;
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE cheques_devueltos ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE cheques_devueltos
    ALTER COLUMN rut_cliente SET NOT NULL,
    ALTER COLUMN monto SET NOT NULL;

ALTER TABLE cheques_devueltos
    ALTER COLUMN fecha_registro SET DEFAULT CURRENT_DATE;

UPDATE cheques_devueltos
SET fecha_registro = COALESCE(fecha_registro, created_at::date)
WHERE fecha_registro IS NULL;

ALTER TABLE cheques_devueltos
    DROP CONSTRAINT IF EXISTS cheques_devueltos_id_cheque_fkey;

ALTER TABLE cheques_devueltos
    ADD CONSTRAINT cheques_devueltos_id_cheque_fkey
    FOREIGN KEY (id_cheque) REFERENCES cheques(id) ON DELETE SET NULL;

ALTER TABLE cheques_devueltos
    DROP CONSTRAINT IF EXISTS cheques_devueltos_rut_cliente_fkey;

ALTER TABLE cheques_devueltos
    ADD CONSTRAINT cheques_devueltos_rut_cliente_fkey
    FOREIGN KEY (rut_cliente) REFERENCES clientes(rut);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cheques_devueltos_id_cheque ON cheques_devueltos(id_cheque);

INSERT INTO cheques_devueltos (
    id_cheque,
    rut_cliente,
    numero_cheque,
    numero_factura,
    monto,
    fecha_registro,
    fecha_cheque,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    c.id,
    c.rut_cliente,
    c.numero_cheque,
    c.numero_factura,
    c.monto,
    CURRENT_DATE,
    NULL,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    NULL
FROM cheques c
WHERE c.deleted_at IS NULL
  AND c.id_estado = 3
ON CONFLICT (id_cheque) DO UPDATE SET
    rut_cliente = EXCLUDED.rut_cliente,
    numero_cheque = EXCLUDED.numero_cheque,
    numero_factura = EXCLUDED.numero_factura,
    monto = EXCLUDED.monto,
    fecha_registro = COALESCE(cheques_devueltos.fecha_registro, EXCLUDED.fecha_registro),
    fecha_cheque = COALESCE(cheques_devueltos.fecha_cheque, EXCLUDED.fecha_cheque),
    updated_at = CURRENT_TIMESTAMP,
    deleted_at = NULL;

CREATE OR REPLACE FUNCTION sync_cheques_devueltos_from_cheques()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;

    IF NEW.deleted_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    IF COALESCE(NEW.id_estado, 0) <> 3 THEN
        RETURN NEW;
    END IF;

    INSERT INTO cheques_devueltos (
        id_cheque,
        rut_cliente,
        numero_cheque,
        numero_factura,
        monto,
        fecha_registro,
        fecha_cheque,
        created_at,
        updated_at,
        deleted_at
    )
    VALUES (
        NEW.id,
        NEW.rut_cliente,
        NEW.numero_cheque,
        NEW.numero_factura,
        NEW.monto,
        CURRENT_DATE,
        NULL,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        NULL
    )
    ON CONFLICT (id_cheque) DO UPDATE SET
        rut_cliente = EXCLUDED.rut_cliente,
        numero_cheque = EXCLUDED.numero_cheque,
        monto = EXCLUDED.monto,
        fecha_registro = COALESCE(cheques_devueltos.fecha_registro, EXCLUDED.fecha_registro),
        fecha_cheque = COALESCE(cheques_devueltos.fecha_cheque, EXCLUDED.fecha_cheque),
        updated_at = CURRENT_TIMESTAMP,
        deleted_at = NULL;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_cheques_devueltos ON cheques;

CREATE TRIGGER trg_sync_cheques_devueltos
AFTER INSERT OR UPDATE OR DELETE ON cheques
FOR EACH ROW
EXECUTE FUNCTION sync_cheques_devueltos_from_cheques();

CREATE OR REPLACE FUNCTION notify_cheques_change()
RETURNS TRIGGER AS $$
DECLARE
    payload TEXT;
BEGIN
    payload := json_build_object(
        'table', TG_TABLE_NAME,
        'operation', TG_OP,
        'id', COALESCE(NEW.id, OLD.id),
        'at', to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')
    )::text;
    PERFORM pg_notify('cheques_changed', payload);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_cheques_changed ON cheques;

CREATE TRIGGER trg_cheques_changed
AFTER INSERT OR UPDATE OR DELETE ON cheques
FOR EACH ROW
EXECUTE FUNCTION notify_cheques_change();

DROP TRIGGER IF EXISTS trg_cheques_devueltos_changed ON cheques_devueltos;

CREATE TRIGGER trg_cheques_devueltos_changed
AFTER INSERT OR UPDATE OR DELETE ON cheques_devueltos
FOR EACH ROW
EXECUTE FUNCTION notify_cheques_change();

DROP TRIGGER IF EXISTS trg_cabinets_changed ON movimientos_cabinets;

CREATE TRIGGER trg_cabinets_changed
AFTER INSERT OR UPDATE OR DELETE ON movimientos_cabinets
FOR EACH ROW
EXECUTE FUNCTION notify_cheques_change();

-- Indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_cheques_rut_cliente ON cheques(rut_cliente);
CREATE INDEX IF NOT EXISTS idx_cheques_id_estado ON cheques(id_estado);
CREATE INDEX IF NOT EXISTS idx_cheques_fecha_cheque_cobrar ON cheques(fecha_cheque_cobrar);
CREATE INDEX IF NOT EXISTS idx_cheques_deleted_at ON cheques(deleted_at);
CREATE INDEX IF NOT EXISTS idx_cheques_devueltos_rut_cliente ON cheques_devueltos(rut_cliente);
CREATE INDEX IF NOT EXISTS idx_cheques_devueltos_deleted_at ON cheques_devueltos(deleted_at);
CREATE INDEX IF NOT EXISTS idx_cheques_devueltos_fecha_saldada ON cheques_devueltos(fecha_saldada);
CREATE INDEX IF NOT EXISTS idx_cheques_devueltos_fecha_registro ON cheques_devueltos(fecha_registro);
CREATE INDEX IF NOT EXISTS idx_cheques_devueltos_fecha_cheque ON cheques_devueltos(fecha_cheque);
CREATE INDEX IF NOT EXISTS idx_movimientos_cabinets_nombre_cliente ON movimientos_cabinets(nombre_cliente);
CREATE INDEX IF NOT EXISTS idx_movimientos_cabinets_fecha_entrada ON movimientos_cabinets(fecha_entrada);
CREATE INDEX IF NOT EXISTS idx_movimientos_cabinets_deleted_at ON movimientos_cabinets(deleted_at);
