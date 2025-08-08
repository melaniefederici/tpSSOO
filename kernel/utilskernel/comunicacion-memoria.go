package utilskernel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func SolicitudInicializarProceso(ip string, puerto int, proceso mensajes.ProcesoAInicializar) bool {
	body, err := json.Marshal(proceso)
	if err != nil {
		log.Printf("Error codificando proceso: %s", err.Error())
		return false
	}

	url := fmt.Sprintf("http://%s:%d/solicitudProcesoAInicializar", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando proceso a Memoria: %s", err.Error())
		return false
	}
	defer resp.Body.Close()

	var respuesta struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		log.Printf("Error al decodificar respuesta de Memoria: %s", err.Error())
		return false
	}

	log.Printf("Respuesta de Memoria: %v", respuesta.OK)
	return respuesta.OK
}

func SolicitudFinalizarProceso(ip string, puerto int, pid int) bool {
	body, err := json.Marshal(map[string]int{"pid": pid})
	if err != nil {
		log.Printf("Error codificando PID: %s", err.Error())
		return false
	}

	url := fmt.Sprintf("http://%s:%d/solicitudFinalizarProceso", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando solicitud de finalización a Memoria: %s", err.Error())
		return false
	}
	defer resp.Body.Close()

	var respuesta struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		log.Printf("Error al decodificar respuesta de Memoria: %s", err.Error())
		return false
	}

	log.Printf("Respuesta de Memoria (finalizar PID %d): %v", pid, respuesta.OK)
	return respuesta.OK
}

func RealizarDumpDeMemoria(proceso *globals.PCB) {
	body, err := json.Marshal(map[string]int{"pid": proceso.PID})
	if err != nil {
		log.Printf("Error codificando DUMP: %s", err.Error())
		FinalizarProceso(proceso)
		return
	}

	url := fmt.Sprintf("http://%s:%d/solicitudDumpMemory", globals.Config.IpMemoria, globals.Config.PuertoMemoria)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando solicitud de DUMP a Memoria: %s", err.Error())
		FinalizarProceso(proceso)
		return
	}
	defer resp.Body.Close()

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error leyendo respuesta de memoria: %s", err.Error())
		FinalizarProceso(proceso)
		return
	}

	log.Printf("Solicitud de DUMP enviada para PID %d", proceso.PID)
	log.Printf("Respuesta de Memoria al DUMP: %s", string(bodyResp))

	var respJSON map[string]string
	err = json.Unmarshal(bodyResp, &respJSON)
	if err != nil || respJSON["status"] != "ok" {
		log.Printf("Dump falló o respuesta inválida, PID %d a EXIT", proceso.PID)
		FinalizarProceso(proceso)
		return
	}

	MoverACola(proceso, globals.ColaREADY)
}

func SolicitudSwapDeProceso(ip string, puerto int, pid int) {
	body, err := json.Marshal(map[string]int{"pid": pid})
	if err != nil {
		log.Printf("Error codificando PID: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/solicitudSwapDeProceso", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando solicitud de finalización a Memoria: %s", err.Error())
	}
	defer resp.Body.Close()

	var respuesta struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		log.Printf("Error al decodificar respuesta de Memoria: %s", err.Error())
	}

	log.Printf("Respuesta de Memoria (SWAPEAR PID %d): %v", pid, respuesta.OK)
}

func SolicitudDesSuspender(ipMemoria string, puertoMemoria int, pid int) bool {
	url := fmt.Sprintf("http://%s:%d/solicitudDesuspender", ipMemoria, puertoMemoria)

	peticion := struct {
		PID int `json:"pid"`
	}{
		PID: pid,
	}

	jsonData, err := json.Marshal(peticion)
	if err != nil {
		log.Printf("Error al serializar pedido de desuspension para PID %d: %s", pid, err)
		return false
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error al enviar pedido de desuspension a Memoria para PID %d: %s", pid, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Respuesta no OK al pedir desuspension a Memoria para PID %d: %s", pid, resp.Status)
		return false
	}

	var respuesta struct {
		Ok bool `json:"ok"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		log.Printf("Error al decodificar respuesta de desuspension de Memoria para PID %d: %s", pid, err)
		return false
	}

	return respuesta.Ok
}
