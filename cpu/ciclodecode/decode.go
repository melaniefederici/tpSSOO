package ciclodecode

import (
	"log"
	"strings"
)

type InstruccionDecodificada struct {
	Tipo               string
	Parametros         []string
	NecesitaTraduccion bool
}

func DecodeInstruccion(instruccion string, parametros []string) InstruccionDecodificada {
	log.Printf("## DECODE - Instrucción: %s | Parámetros: %v", instruccion, parametros)

	inst := strings.ToUpper(instruccion) //por si esta en minuscula

	var instruccionDecodificada InstruccionDecodificada
	instruccionDecodificada.Parametros = parametros

	switch inst {
	case "READ", "WRITE":
		instruccionDecodificada.NecesitaTraduccion = true
		instruccionDecodificada.Tipo = inst

	case "IO", "INIT_PROC", "NOOP", "DUMP_MEMORY", "EXIT", "GOTO":
		instruccionDecodificada.NecesitaTraduccion = false
		instruccionDecodificada.Tipo = inst

	default:
		log.Printf("Instrucción desconocida: %s", instruccion)
		instruccionDecodificada.Tipo = "DESCONOCIDA"
	}

	return instruccionDecodificada
}
