package ciclofetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	//"github.com/sisoputnfrba/tp-golang/cpu/mmu"
)

type PeticionMemoria struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type RespuestaMemoria struct {
	Instruccion string   `json:"instruccion"`
	Parametros  []string `json:"parametros"`
}

func FetchInstruccion(pid int, pc int) (string, []string) {
	fmt.Printf("Solicitando instrucción a Memoria: PID=%d | PC=%d\n", pid, pc)

	// direccionFisica := mmu.TraducirDireccion(pc, pid)
	// if direccionFisica == -1 {
	//     fmt.Println("Error: No se pudo traducir la dirección lógica.")
	//     return "", nil
	// }

	peticion := PeticionMemoria{
		PID: pid,
		PC:  pc, // PC: direccionFisica
	}

	body, err := json.Marshal(peticion)
	if err != nil {
		fmt.Println("Error al codificar JSON:", err)
		return "", nil
	}

	url := fmt.Sprintf("http://%s:%d/solicitudInstruccion", globals.Config.IPMemoria, globals.Config.PuertoMemoria)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))

	if err != nil {
		fmt.Println("Error al solicitar instrucción a Memoria:", err)
		return "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Memoria respondió con un error:", resp.Status)
		return "", nil
	}

	var respuesta RespuestaMemoria
	err = json.NewDecoder(resp.Body).Decode(&respuesta)
	if err != nil {
		fmt.Println("Error al decodificar respuesta de Memoria:", err)
		return "", nil
	}

	if respuesta.Instruccion == "" {
		fmt.Println("No se recibió una instrucción válida desde Memoria.")
		return "", nil
	}

	fmt.Printf("Instrucción recibida: %s | Parámetros: %v\n", respuesta.Instruccion, respuesta.Parametros)
	return respuesta.Instruccion, respuesta.Parametros
}
