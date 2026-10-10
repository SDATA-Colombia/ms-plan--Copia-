package cache

import "time"

const (

	// Sin expiración.
	TTLForever time.Duration = 0

	// Datos que cambian constantemente.
	//
	// Ejemplo:
	// - Stock
	// - Dashboard
	// - Notificaciones
	TTLRealtime = 5 * time.Minute

	// Datos de negocio.
	//
	// Ejemplo:
	// - Usuarios
	// - Clientes
	// - Productos
	// - Colores
	TTLStandard = 30 * time.Minute

	// Catálogos.
	//
	// Ejemplo:
	// - Países
	// - Ciudades
	// - Departamentos
	TTLLong = time.Hour

	// Configuración prácticamente estática.
	TTLDaily = 24 * time.Hour
)
