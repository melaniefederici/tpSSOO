package utilscpu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/cache"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/tlb"
)

type Recepcion struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

// para cpu
func EsperarRespuesta() Recepcion {
	puerto := fmt.Sprintf(":%d", globals.Config.PuertoCPUDispatch)
	ln, err := net.Listen("tcp", puerto)
	if err != nil {
		log.Fatalf("Error al escuchar en el puerto %s: %v", puerto, err)
	}
	defer ln.Close()

	log.Printf("Esperando conexión del Kernel en el puerto %s...", puerto)

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("Error al aceptar conexión: %v", err)
	}
	defer conn.Close()

	var recepcion Recepcion
	err = json.NewDecoder(conn).Decode(&recepcion)
	if err != nil {
		log.Fatalf("Error al decodificar JSON del Kernel: %v", err)
	}

	log.Printf("Recibiendo del Kernel: PID=%d, PC=%d", recepcion.PID, recepcion.PC)
	return recepcion
}

// para execute
func ValidarParametros(inst string, esperados int, params []string) bool {
	if len(params) != esperados {
		log.Printf("Error: %s necesita %d parámetros.", inst, esperados)
		return false
	}
	return true
}

func FinalizarProceso(pid int) {
	tlb.LimpiarTLB(pid)
	LimpiarBitmapYTablas(pid)
	cache.LimpiarProceso(pid)
}

// Para execute
func PedirPaginaAMemoria(pid int, dirFisica int) []byte {
	url := fmt.Sprintf("http://%s:%d/solicitudLecturaPagina", globals.Config.IPMemoria, globals.Config.PuertoMemoria)

	type PeticionLecturaPagina struct {
		PID       int `json:"pid"`
		DirFisica int `json:"dir_fisica"`
	}

	req := PeticionLecturaPagina{
		PID:       pid,
		DirFisica: dirFisica,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error al leer página de memoria: %v", err)
		return []byte{}
	}
	defer resp.Body.Close()

	var contenido []byte
	_ = json.NewDecoder(resp.Body).Decode(&contenido)
	return contenido
}
