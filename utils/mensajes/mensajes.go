package mensajes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Motivo string

const (
	MotivoIO         Motivo = "IO"
	MotivoExit       Motivo = "EXIT"
	MotivoDumpMemory Motivo = "DUMP_MEMORY"
	MotivoDesalojo   Motivo = "DESALOJO"
)

type Mensaje struct {
	Mensaje string `json:"mensaje"`
}

type Paquete struct {
	Valores []string `json:"valores"`
}

type ProcesoAInicializar struct {
	PID           int    `json:"pid"`
	NombreArchivo string `json:"archivo"`
	Tamanio       int    `json:"tamanio"`
}

type ProcesoAEjecutar struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type ResultadoCPU struct {
	PID              int      `json:"pid"`
	PC               int      `json:"pc"`
	Motivo           Motivo   `json:"motivo"`
	Args             []string `json:"args"`
	IdentificadorCPU int      `json:"identificador"`
}

func RecibirMensaje(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var mensaje Mensaje
	err := decoder.Decode(&mensaje)
	if err != nil {
		log.Printf("Error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	log.Println("Me llego un mensaje de un cliente")
	log.Printf("%+v\n", mensaje)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func EnviarMensaje(ip string, puerto int, mensajeTxt string) {
	mensaje := Mensaje{Mensaje: mensajeTxt}
	body, err := json.Marshal(mensaje)
	if err != nil {
		log.Printf("error codificando mensaje: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/mensaje", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensaje a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s", resp.Status)
}

func GenerarYEnviarPaquete(ip string, puerto int, valores []string) {
	paquete := Paquete{}
	// Leemos y cargamos el paquete
	paquete.Valores = valores
	log.Printf("paquete a enviar: %+v", paquete)
	// Enviamos el paqute
	EnviarPaquete(ip, puerto, paquete)
}

func EnviarPaquete(ip string, puerto int, paquete Paquete) {
	body, err := json.Marshal(paquete)
	if err != nil {
		log.Printf("error codificando mensajes: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/paquetes", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensajes a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s", resp.Status)
}

func RecibirPaquetes(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var paquete Paquete
	err := decoder.Decode(&paquete)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	log.Println("me llego un paquete de un cliente")
	log.Printf("%+v\n", paquete)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func EnviarProcesoAInicializar(ip string, puerto int, proceso ProcesoAInicializar) {
	body, err := json.Marshal(proceso)
	if err != nil {
		log.Printf("error codificando mensajes: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/procesoAInicializar", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensajes a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s", resp.Status)
}

func RecibirProcesoAInicializar(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var proceso ProcesoAInicializar
	err := decoder.Decode(&proceso)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	log.Printf("Me llego el proceso %d para inicializarlo", proceso.PID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
