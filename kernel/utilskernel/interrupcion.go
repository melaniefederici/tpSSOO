package utilskernel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
)

func VerificarDesalojoPorSRT(nuevo *globals.PCB) bool {
	desalojar := false
	for _, cpu := range globals.CPUsConectadas {
		if cpu.Ocupado && cpu.ProcesoEjecutando != nil {
			ejecutando := cpu.ProcesoEjecutando

			// Calcular cuánto ejecutó realmente
			tiempoEjecutado := time.Now().Sub(ejecutando.InicioExec)

			// Calcular tiempo restante estimado
			ejecutadoMs := float64(tiempoEjecutado.Milliseconds())

			tiempoRestante := ejecutando.UltimaEstimacion - ejecutadoMs

			if nuevo.EstimacionRafaga < tiempoRestante && ejecutando.InterrupcionEnviada == false {
				log.Printf("###### SRT - Se interrumpe la CPU %d: PID %d será desalojado por PID %d",
					cpu.Identificador, ejecutando.PID, nuevo.PID)
				log.Printf("## (%d) - Desalojado por algoritmo SJF/SRT", ejecutando.PID)

				url := fmt.Sprintf("http://%s:%d/interrupcion", cpu.IP, cpu.PuertoInterrupt)
				body, _ := json.Marshal(map[string]int{"pid": ejecutando.PID})

				resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
				if err != nil {
					log.Printf("Error al enviar interrupción SRT: %s", err.Error())
					return false
				}
				defer resp.Body.Close()
				log.Printf("Se envió interrupción por SRT a la CPU %d", cpu.Identificador)
				ejecutando.InterrupcionEnviada = true
				desalojar = true
			}
		}
	}
	return desalojar
}
