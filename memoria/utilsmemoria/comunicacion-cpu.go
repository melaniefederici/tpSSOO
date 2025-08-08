package utilsmemoria

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func RespuestaInstruccion(w http.ResponseWriter, r *http.Request) {
	var peticion globals.PeticionInstruccion

	err := json.NewDecoder(r.Body).Decode(&peticion)
	if err != nil {
		http.Error(w, "Error al decodificar petición", http.StatusBadRequest)
		return
	}

	instruccion, parametros := BuscarInstruccion(peticion.PID, peticion.PC)
	if instruccion == "" || parametros == nil {
		http.Error(w, "No se pudo obtener la instrucción", http.StatusNotFound)
		return
	}

	log.Printf("## PID: %d - Obtener instruccion: %d - Instruccion: %s, %v", peticion.PID, peticion.PC, instruccion, parametros)

	respuesta := globals.RespuestaInstruccion{
		Instruccion: instruccion,
		Parametros:  parametros,
	}

	time.Sleep(time.Duration(globals.Config.MemoryDelay) * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

func RespuestaLectura(w http.ResponseWriter, r *http.Request) {
	var peticion globals.PeticionLectura
	err := json.NewDecoder(r.Body).Decode(&peticion)
	if err != nil {
		http.Error(w, "Error al decodificar petición de lectura", http.StatusBadRequest)
		return
	}

	respuesta := LeerContenido(peticion.PID, peticion.DirFisica, peticion.Tamanio)

	log.Printf("## PID: %d - Lectura - Dir.Fisica: %d - Tamanio: %d", peticion.PID, peticion.DirFisica, peticion.Tamanio)
	json.NewEncoder(w).Encode(respuesta)
}

func RespuestaEscritura(w http.ResponseWriter, r *http.Request) {
	var peticion globals.PeticionEscritura
	if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
		http.Error(w, "Error al decodificar peticion de escritura", http.StatusBadRequest)
		return
	}

	EscribirContenido(peticion.PID, peticion.DirFisica, peticion.Cadena)

	log.Printf("## PID: %d - Escritura - Dir.Fisica: %d - Tamanio: %d", peticion.PID, peticion.DirFisica, len(peticion.Cadena))
	contenidoEscrito := LeerContenido(peticion.PID, peticion.DirFisica, len(peticion.Cadena))
	log.Printf("Escribi este contenido: %v", contenidoEscrito)

	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func RespuestaLecturaPagina(w http.ResponseWriter, r *http.Request) {
	var peticion globals.PeticionLecturaPagina
	if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
		http.Error(w, "Error al decodificar petición de lectura de pagina", http.StatusBadRequest)
		return
	}

	respuesta := LeerPaginaCompleta(peticion.PID, peticion.DirFisica)
	log.Printf("Le mando a CPU el array de bytes: %v", respuesta)
	log.Printf("## PID: %d - Lectura - Dir.Fisica: %d - Tamanio: %d", peticion.PID, peticion.DirFisica, globals.Config.PageSize)
	json.NewEncoder(w).Encode(respuesta)
}

func RespuestaActualizarPagina(w http.ResponseWriter, r *http.Request) {
	log.Printf("Me llego la peticion de ACTUALIZAR pagina")
	var peticion globals.PeticionActualizarPagina
	if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
		http.Error(w, "Error al decodificar petición de actualizacion de pagina", http.StatusBadRequest)
		return
	}

	ActualizarPaginaCompleta(peticion.PID, peticion.DirFisica, peticion.Cadena)

	log.Printf("## PID: %d - Escritura - Dir.Fisica: %d - Tamanio: %d", peticion.PID, peticion.DirFisica, globals.Config.PageSize)

	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func RespuestaObtenerMarco(w http.ResponseWriter, r *http.Request) {
	var datos struct {
		PID       int `json:"pid"`
		NroPagina int `json:"nroPagina"`
	}

	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	marco := ObtenerMarco(datos.PID, datos.NroPagina)

	respuesta := map[string]int{"marco": marco}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

func ObtenerConfigHandler(w http.ResponseWriter, r *http.Request) {
	config := struct {
		TamPagina         int `json:"tam_pagina"`
		CantEntradasTabla int `json:"cant_entradas_tabla"`
		CantNiveles       int `json:"cant_niveles"`
	}{
		TamPagina:         globals.Config.PageSize,
		CantEntradasTabla: globals.Config.EntriesPerPage,
		CantNiveles:       globals.Config.NumberOfLevels,
	}

	json.NewEncoder(w).Encode(config)
}
