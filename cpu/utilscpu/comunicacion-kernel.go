package utilscpu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

// mismo problema en cpu
func RegistrarEnKernel(kernelIP string, kernelPuerto int, identificador int) {

	cpu := struct {
		IP              string `json:"ip_cpu"`
		PuertoDispatch  int    `json:"puerto_dispatch"`
		PuertoInterrupt int    `json:"puerto_interrupt"`
		Identificador   int    `json:"identificador_cpu"`
	}{
		IP:              globals.Config.IPCPU,
		PuertoDispatch:  globals.Config.PuertoCPUDispatch,
		PuertoInterrupt: globals.Config.PuertoCPUInterrupt,
		Identificador:   identificador,
	}

	url := fmt.Sprintf("http://%s:%d/registrarCPU", kernelIP, kernelPuerto)

	body, err := json.Marshal(cpu)
	if err != nil {
		log.Fatalf("Error al serializar CPU: %s", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error al registrar CPU en Kernel: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Kernel devolvió error al registrar CPU: %d", resp.StatusCode)
	}

	log.Println("CPU registrada correctamente en el Kernel")
}

func EnviarResultadoCPU(kernelIP string, kernelPuerto int, resultado mensajes.ResultadoCPU) {
	url := fmt.Sprintf("http://%s:%d/recibirResultado", kernelIP, kernelPuerto)
	body, err := json.Marshal(resultado)
	if err != nil {
		log.Fatalf("Error al serializar el resultado: %s", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error al enviar resultado a Kernel: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Kernel devolvió error recibir resultado de CPU: %d", resp.StatusCode)
	}

	log.Println("Resultado enviado correctamente al Kernel")
}

func EnviarInitProc(ip string, puerto int, datos []string) {
	url := fmt.Sprintf("http://%s:%d/manejarINITPROC", ip, puerto)
	body, _ := json.Marshal(datos)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error al enviar resultado a Kernel: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Kernel respondió con error: %s", resp.Status)
	}
}

func RecibirProcesoAEjecutar(w http.ResponseWriter, r *http.Request) { //agregar a handlers del CPU en puerto dispatch "/solicitudProcesoAInicializar"
	var proceso mensajes.ProcesoAEjecutar
	if err := json.NewDecoder(r.Body).Decode(&proceso); err != nil {
		log.Printf("Error al recibir proceso: %s", err.Error())
		return
	}
	defer r.Body.Close()

	log.Printf("CPU recibió proceso: %+v", proceso)

	respuesta := struct {
		OK bool `json:"ok"`
	}{OK: true}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)

	globals.CanalProcesoAEjecutar <- proceso
}
