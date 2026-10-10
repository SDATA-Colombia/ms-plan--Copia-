package cache

import (
	"fmt"
	"strings"
)

// Separator utilizado para construir todas las llaves.
const Separator = ":"

// Key construye una llave uniendo todas sus partes.
//
// Ejemplos:
//
//	Key("color", "id", id)
//	=> color:id:685b6c2f...
//
//	Key("invoice", "page", hash)
//	=> invoice:page:a87dc12...
func Key(parts ...any) string {

	values := make([]string, 0, len(parts))

	for _, part := range parts {
		values = append(values, fmt.Sprint(part))
	}

	return strings.Join(values, Separator)
}

// VersionKey devuelve la llave donde se almacena la versión
// de un namespace.
//
// Ejemplo:
//
//	VersionKey("color", "company", companyID)
//
// Resultado:
//
//	color:company:10:version
func VersionKey(parts ...any) string {

	values := append(parts, "version")

	return Key(values...)
}

// CacheKey devuelve una llave versionada.
//
// Ejemplo:
//
//	CacheKey(
//	    version,
//	    "color",
//	    "company",
//	    companyID,
//	    "list",
//	    hash,
//	)
//
// Resultado:
//
//	color:company:10:v:15:list:af87c...
func CacheKey(
	version int64,
	parts ...any,
) string {

	values := make([]any, 0, len(parts)+2)

	values = append(values, parts...)
	values = append(values, "v")
	values = append(values, version)

	return Key(values...)
}
