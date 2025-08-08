package planificadores

import (
	"sort"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/utilskernel"
)

func PlanificarCP() {
	for {
		<-globals.HayProcesoEnReady
	//planificacion:
		for {
			select {
			case cpu := <-globals.CPUDisponibles:
				globals.ColaREADY.MutexCola.Lock()
				if len(globals.ColaREADY.Cola) == 0 {
					globals.ColaREADY.MutexCola.Unlock()
					globals.CPUDisponibles <- cpu // devuelvo el cpu a la pool de cpus disponibles
					time.Sleep(100 * time.Millisecond)
					continue
					//break planificacion
				}

				switch globals.Config.PlanificacionCP {
				case "FIFO":
					// hacer nada
				case "SJF", "SRT":
					// Ordenar por EstimacionRafaga (de menor a mayor)
					sort.Slice(globals.ColaREADY.Cola, func(i, j int) bool {
						return globals.ColaREADY.Cola[i].EstimacionRafaga < globals.ColaREADY.Cola[j].EstimacionRafaga
					})
				}
				proceso := globals.ColaREADY.Cola[0]
				globals.ColaREADY.MutexCola.Unlock()
				utilskernel.MoverACola(proceso, globals.ColaEXEC)
				proceso.UltimaEstimacion = proceso.EstimacionRafaga
				proceso.InicioExec = time.Now()
				procesoAEjecutar := utilskernel.ConvertirAProcesoAEjecutar(proceso)
				utilskernel.EnviarProcesoACpu(cpu.IP, cpu.PuertoDispatch, procesoAEjecutar)
				//utilskernel.ImprimirEstadoColas() // 												PARA TESTEOS
				cpu.ProcesoEjecutando = proceso
				cpu.Ocupado = true

			default:
				if globals.Config.PlanificacionCP == "SRT" {
					globals.ColaREADY.MutexCola.Lock()
					if len(globals.ColaREADY.Cola) > 0 {
						sort.Slice(globals.ColaREADY.Cola, func(i, j int) bool {
							return globals.ColaREADY.Cola[i].EstimacionRafaga < globals.ColaREADY.Cola[j].EstimacionRafaga
						})
						proceso := globals.ColaREADY.Cola[0]
						globals.ColaREADY.MutexCola.Unlock()
						if utilskernel.VerificarDesalojoPorSRT(proceso) {
							// ESPERAR A QUE SE DESALOJE
							<-globals.DesalojoHecho
							utilskernel.MoverACola(proceso, globals.ColaEXEC)
							proceso.UltimaEstimacion = proceso.EstimacionRafaga
							proceso.InicioExec = time.Now()
							procesoAEjecutar := utilskernel.ConvertirAProcesoAEjecutar(proceso)
							cpu := <-globals.CPUDisponibles
							utilskernel.EnviarProcesoACpu(cpu.IP, cpu.PuertoDispatch, procesoAEjecutar)
							cpu.ProcesoEjecutando = proceso
							cpu.Ocupado = true
						//} else {
						//	break planificacion
						}
					} else {
						globals.ColaREADY.MutexCola.Unlock()
						//break planificacion
					}
					//si no es SRT sigue y vuelve al bucle
				//} else {
				//	break planificacion
				}
			}
		}
	}
}
