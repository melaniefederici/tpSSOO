package utilscpu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

func SolicitudEscritura(ip string, puerto int, solicitud globals.PeticionEscritura) {
	body, err := json.Marshal(solicitud)
	if err != nil {
		log.Printf("Error codificando la solicitud de escritura: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/solicitudEscritura", ip, puerto)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error al enviar solicitud de escritura a Memoria: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	var respuesta struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		log.Printf("Error al decodificar respuesta de Memoria: %s", err.Error())
		return
	}

	log.Printf("Respuesta de Memoria a escritura: %v", respuesta.OK)
}

func ObtenerConfigMemoria() {
	url := fmt.Sprintf("http://%s:%d/solicitudConfig", globals.Config.IPMemoria, globals.Config.PuertoMemoria)

	log.Printf("La url que mando es: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error al hacer GET a Memoria: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Memoria respondió con estado: %d", resp.StatusCode)
	}

	var config struct {
		TamPagina         int `json:"tam_pagina"`
		CantEntradasTabla int `json:"cant_entradas_tabla"`
		CantNiveles       int `json:"cant_niveles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		log.Fatalf("Error al decodificar respuesta de Memoria: %s", err)
	}

	globals.Config.TamanioPagina = config.TamPagina
	globals.Config.CantEntradasTabla = config.CantEntradasTabla
	globals.Config.CantNiveles = config.CantNiveles

	log.Printf("Config de Memoria recibida: Tamaño de página: %d, Entradas por tabla: %d, Niveles: %d",
		config.TamPagina, config.CantEntradasTabla, config.CantNiveles)
}
