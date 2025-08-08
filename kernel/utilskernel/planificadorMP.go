package utilskernel

import (
	"log"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
)

func PlanificarMP(PID int) {
	proceso := BuscarPCBPorPID(PID)
	if proceso == nil {
		log.Printf("[PlanificarMP] Proceso %d no encontrado.", PID)
		return
	}

	proceso.MutexProc.Lock()

	if proceso.Estado != globals.EstadoBlocked {
		log.Printf("[PlanificarMP] Proceso %d ya no está bloqueado, está en <%s>. No se suspende.", PID, proceso.Estado)
		return
	}
	proceso.MutexProc.Unlock()

	log.Printf("[PlanificarMP] Suspendiendo al proceso %d", PID)

	proceso.TimerSuspension = nil // Se ejecutó el timer

	SolicitudSwapDeProceso(globals.Config.IpMemoria, globals.Config.PuertoMemoria, PID)
	MoverACola(proceso, globals.ColaSuspBlock)
	log.Printf("Ya mande señal de memoria liberada")

	globals.SenialMemoriaLiberada <- struct{}{}
	log.Printf("Ya mande señal de memoria liberada")
}
