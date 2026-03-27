export namespace main {
	
	export class Cheque {
	    id: number;
	    numero_cheque?: string;
	    rut_cliente: string;
	    nombre_cliente: string;
	    monto: number;
	    fecha_recepcion?: string;
	    fecha_deposito?: string;
	    fecha_cheque_cobrar?: string;
	    banco_cheque?: string;
	    numero_factura?: string;
	    condiciones_pago?: string;
	    observaciones?: string;
	    id_estado?: number;
	    estado_nombre?: string;
	    nombre_vendedor?: string;
	
	    static createFrom(source: any = {}) {
	        return new Cheque(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.numero_cheque = source["numero_cheque"];
	        this.rut_cliente = source["rut_cliente"];
	        this.nombre_cliente = source["nombre_cliente"];
	        this.monto = source["monto"];
	        this.fecha_recepcion = source["fecha_recepcion"];
	        this.fecha_deposito = source["fecha_deposito"];
	        this.fecha_cheque_cobrar = source["fecha_cheque_cobrar"];
	        this.banco_cheque = source["banco_cheque"];
	        this.numero_factura = source["numero_factura"];
	        this.condiciones_pago = source["condiciones_pago"];
	        this.observaciones = source["observaciones"];
	        this.id_estado = source["id_estado"];
	        this.estado_nombre = source["estado_nombre"];
	        this.nombre_vendedor = source["nombre_vendedor"];
	    }
	}
	export class ChequeDevuelto {
	    id: number;
	    id_cheque?: number;
	    numero_cheque?: string;
	    numero_factura?: string;
	    monto: number;
	    rut_cliente: string;
	    nombre_cliente: string;
	    plaza?: string;
	    fecha_registro?: string;
	    fecha_cheque?: string;
	    tipo_pago?: string;
	    motivo?: string;
	    fecha_saldada?: string;
	    comentario?: string;
	    estado_pago: string;
	    created_at?: string;
	    updated_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new ChequeDevuelto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.id_cheque = source["id_cheque"];
	        this.numero_cheque = source["numero_cheque"];
	        this.numero_factura = source["numero_factura"];
	        this.monto = source["monto"];
	        this.rut_cliente = source["rut_cliente"];
	        this.nombre_cliente = source["nombre_cliente"];
	        this.plaza = source["plaza"];
	        this.fecha_registro = source["fecha_registro"];
	        this.fecha_cheque = source["fecha_cheque"];
	        this.tipo_pago = source["tipo_pago"];
	        this.motivo = source["motivo"];
	        this.fecha_saldada = source["fecha_saldada"];
	        this.comentario = source["comentario"];
	        this.estado_pago = source["estado_pago"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ChequeDevueltoInput {
	    id_cheque?: number;
	    rut_cliente: string;
	    nombre_cliente: string;
	    numero_cheque: string;
	    numero_factura: string;
	    monto: number;
	    plaza: string;
	    fecha_registro: string;
	    fecha_cheque: string;
	    tipo_pago: string;
	    motivo: string;
	    fecha_saldada: string;
	    comentario: string;
	
	    static createFrom(source: any = {}) {
	        return new ChequeDevueltoInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id_cheque = source["id_cheque"];
	        this.rut_cliente = source["rut_cliente"];
	        this.nombre_cliente = source["nombre_cliente"];
	        this.numero_cheque = source["numero_cheque"];
	        this.numero_factura = source["numero_factura"];
	        this.monto = source["monto"];
	        this.plaza = source["plaza"];
	        this.fecha_registro = source["fecha_registro"];
	        this.fecha_cheque = source["fecha_cheque"];
	        this.tipo_pago = source["tipo_pago"];
	        this.motivo = source["motivo"];
	        this.fecha_saldada = source["fecha_saldada"];
	        this.comentario = source["comentario"];
	    }
	}
	export class ChequeInput {
	    numero_cheque: string;
	    rut_cliente: string;
	    nombre_cliente: string;
	    monto: number;
	    fecha_recepcion: string;
	    fecha_deposito: string;
	    fecha_cheque_cobrar: string;
	    banco_cheque: string;
	    numero_factura: string;
	    condiciones_pago: string;
	    observaciones: string;
	    id_estado: number;
	
	    static createFrom(source: any = {}) {
	        return new ChequeInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.numero_cheque = source["numero_cheque"];
	        this.rut_cliente = source["rut_cliente"];
	        this.nombre_cliente = source["nombre_cliente"];
	        this.monto = source["monto"];
	        this.fecha_recepcion = source["fecha_recepcion"];
	        this.fecha_deposito = source["fecha_deposito"];
	        this.fecha_cheque_cobrar = source["fecha_cheque_cobrar"];
	        this.banco_cheque = source["banco_cheque"];
	        this.numero_factura = source["numero_factura"];
	        this.condiciones_pago = source["condiciones_pago"];
	        this.observaciones = source["observaciones"];
	        this.id_estado = source["id_estado"];
	    }
	}
	export class Cliente {
	    rut: string;
	    razon_social: string;
	    id_vendedor?: number;
	    zona?: string;
	
	    static createFrom(source: any = {}) {
	        return new Cliente(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rut = source["rut"];
	        this.razon_social = source["razon_social"];
	        this.id_vendedor = source["id_vendedor"];
	        this.zona = source["zona"];
	    }
	}
	export class ClienteMoroso {
	    rut: string;
	    razon_social: string;
	    nombre_vendedor?: string;
	    total_deuda: number;
	    cheques_vencidos: number;
	
	    static createFrom(source: any = {}) {
	        return new ClienteMoroso(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rut = source["rut"];
	        this.razon_social = source["razon_social"];
	        this.nombre_vendedor = source["nombre_vendedor"];
	        this.total_deuda = source["total_deuda"];
	        this.cheques_vencidos = source["cheques_vencidos"];
	    }
	}
	export class EstadoCheque {
	    id: number;
	    nombre: string;
	
	    static createFrom(source: any = {}) {
	        return new EstadoCheque(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nombre = source["nombre"];
	    }
	}
	export class ExportResult {
	    file_path: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file_path = source["file_path"];
	    }
	}
	export class MovimientoCabinet {
	    id: number;
	    codigo_movimiento: string;
	    nombre_cliente: string;
	    direccion?: string;
	    localidad?: string;
	    cantidad_cabinets: number;
	    descripcion?: string;
	    fecha_entrada?: string;
	    fecha_salida?: string;
	    valor: number;
	    created_at?: string;
	    updated_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new MovimientoCabinet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.codigo_movimiento = source["codigo_movimiento"];
	        this.nombre_cliente = source["nombre_cliente"];
	        this.direccion = source["direccion"];
	        this.localidad = source["localidad"];
	        this.cantidad_cabinets = source["cantidad_cabinets"];
	        this.descripcion = source["descripcion"];
	        this.fecha_entrada = source["fecha_entrada"];
	        this.fecha_salida = source["fecha_salida"];
	        this.valor = source["valor"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class MovimientoCabinetInput {
	    nombre_cliente: string;
	    direccion: string;
	    localidad: string;
	    cantidad_cabinets: number;
	    descripcion: string;
	    codigo_movimiento: string;
	    fecha_entrada: string;
	    fecha_salida: string;
	
	    static createFrom(source: any = {}) {
	        return new MovimientoCabinetInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nombre_cliente = source["nombre_cliente"];
	        this.direccion = source["direccion"];
	        this.localidad = source["localidad"];
	        this.cantidad_cabinets = source["cantidad_cabinets"];
	        this.descripcion = source["descripcion"];
	        this.codigo_movimiento = source["codigo_movimiento"];
	        this.fecha_entrada = source["fecha_entrada"];
	        this.fecha_salida = source["fecha_salida"];
	    }
	}
	export class Vendedor {
	    id: number;
	    nombre_completo: string;
	    zona?: string;
	
	    static createFrom(source: any = {}) {
	        return new Vendedor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nombre_completo = source["nombre_completo"];
	        this.zona = source["zona"];
	    }
	}

}

