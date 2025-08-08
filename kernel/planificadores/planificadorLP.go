package planificadores

import (
	"bufio"
	"log"
	"os"
	"sort"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/utilskernel"
)

func InicioPlanificadorLP() bool {
	reader := bufio.NewReader(os.Stdin)
	log.Println("Presione ENTER para iniciar la planificacion a largo plazo")
	for {
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]
		if text == "" {
			return true
		} else {
			log.Println("Tecla incorrecta. Presione ENTER para iniciar la planificacion a largo plazo")
		}
	}
}

func PlanificarLP() {
	if InicioPlanificadorLP() {
		switch globals.Config.PlanificacionLP {
		case "FIFO":
			log.Printf("Se inició la planificacion a largo plazo por %s", globals.Config.PlanificacionLP)
			for {
				globals.ColaNEW.MutexCola.Lock()
				globals.ColaSuspReady.MutexCola.Lock()
				if len(globals.ColaSuspReady.Cola) == 0 && len(globals.ColaNEW.Cola) == 0 {
					globals.ColaNEW.MutexCola.Unlock()
					globals.ColaSuspReady.MutexCola.Unlock()
					time.Sleep(50 * time.Millisecond)
					continue
				}
				if len(globals.ColaSuspReady.Cola) == 0 {
					proceso := globals.ColaNEW.Cola[0]
					globals.ColaNEW.MutexCola.Unlock()
					globals.ColaSuspReady.MutexCola.Unlock()
					procesoAInicializar := utilskernel.ConvertirAProcesoAInicializar(proceso)
					if utilskernel.SolicitudInicializarProceso(globals.Config.IpMemoria, globals.Config.PuertoMemoria, procesoAInicializar) {
						utilskernel.MoverACola(proceso, globals.ColaREADY)
					} else {
						//go intentarSuspenderBloqueados()
						select {
						case <-globals.SenialMemoriaLiberada:
						}
					}
				} else {
					globals.ColaNEW.MutexCola.Unlock()
					proceso := globals.ColaSuspReady.Cola[0]
					globals.ColaSuspReady.MutexCola.Unlock()
					procesoAInicializar := utilskernel.ConvertirAProcesoAInicializar(proceso)
					if utilskernel.SolicitudDesSuspender(globals.Config.IpMemoria, globals.Config.PuertoMemoria, procesoAInicializar.PID) {
						utilskernel.MoverACola(proceso, globals.ColaREADY)
					} else {
						//go intentarSuspenderBloqueados()
						<-globals.SenialMemoriaLiberada
					}
				}
			}
		case "PMCP":
			log.Printf("Se inició la planificacion a largo plazo por %s", globals.Config.PlanificacionLP)
			for {
				globals.ColaNEW.MutexCola.Lock()
				globals.ColaSuspReady.MutexCola.Lock()

				if len(globals.ColaNEW.Cola) == 0 && len(globals.ColaSuspReady.Cola) == 0 {
					globals.ColaNEW.MutexCola.Unlock()
					globals.ColaSuspReady.MutexCola.Unlock()
					time.Sleep(50 * time.Millisecond)
					continue
				}

				if len(globals.ColaSuspReady.Cola) > 0 {
					// Prioridad: SuspReady
					sort.Slice(globals.ColaSuspReady.Cola, func(i, j int) bool {
						return globals.ColaSuspReady.Cola[i].Tamanio < globals.ColaSuspReady.Cola[j].Tamanio
					})
					proceso := globals.ColaSuspReady.Cola[0]
					globals.ColaSuspReady.MutexCola.Unlock()
					globals.ColaNEW.MutexCola.Unlock()

					procesoAInicializar := utilskernel.ConvertirAProcesoAInicializar(proceso)

					if utilskernel.SolicitudDesSuspender(globals.Config.IpMemoria, globals.Config.PuertoMemoria, procesoAInicializar.PID) {
						utilskernel.MoverACola(proceso, globals.ColaREADY)
					} else {
						<-globals.SenialMemoriaLiberada
					}

				} else {
					// Si no hay en SuspReady, se hace PMCP sobre NEW
					sort.Slice(globals.ColaNEW.Cola, func(i, j int) bool {
						return float64(globals.ColaNEW.Cola[i].Tamanio) < float64(globals.ColaNEW.Cola[j].Tamanio)
					})
					proceso := globals.ColaNEW.Cola[0]
					globals.ColaNEW.MutexCola.Unlock()
					globals.ColaSuspReady.MutexCola.Unlock()

					procesoAInicializar := utilskernel.ConvertirAProcesoAInicializar(proceso)

					if utilskernel.SolicitudInicializarProceso(globals.Config.IpMemoria, globals.Config.PuertoMemoria, procesoAInicializar) {
						utilskernel.MoverACola(proceso, globals.ColaREADY)
					} else {
						<-globals.SenialMemoriaLiberada
					}
				}
			}
		}
	}
}

/*func intentarSuspenderBloqueados() {
	globals.ColaBLOCK.MutexCola.Lock()
	defer globals.ColaBLOCK.MutexCola.Unlock()

	for _, proc := range globals.ColaBLOCK.Cola {
		go utilskernel.PlanificarMP(proc.PID)
	}
}
*/
