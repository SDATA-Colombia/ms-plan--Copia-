package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Hash genera un hash SHA-256 determinístico de cualquier estructura.
//
// Se utiliza principalmente para construir llaves de caché cuando una
// consulta depende de filtros, paginación o parámetros complejos.
//
// Ejemplo:
//
//	filter := bson.M{
//	    "status": true,
//	    "company": companyID,
//	}
//
//	hash, _ := cache.Hash(filter)
//
//	key := cache.Key("color", "list", hash)
func Hash(v any) (string, error) {

	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:]), nil
}
