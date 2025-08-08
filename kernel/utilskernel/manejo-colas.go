package utilskernel

import (
	"log"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
)

func MoverACola(proceso *globals.PCB, destino *globals.Cola) {
	proceso.MutexProc.Lock()
	defer proceso.MutexProc.Unlock()

	estadoAnterior := proceso.Estado
	tiempoEnEstado := time.Since(proceso.InicioEstado).Milliseconds()

	proceso.MT[estadoAnterior] += int(tiempoEnEstado)
	proceso.ME[destino.Estado] += 1

	SacarDeColaPorEstado(proceso)

	proceso.Estado = destino.Estado
	proceso.InicioEstado = time.Now()

	destino.MutexCola.Lock()
	defer destino.MutexCola.Unlock()

	destino.Cola = append(destino.Cola, proceso)

	log.Printf("## (%d) Pasa del estado <%s> al estado <%s>", proceso.PID, estadoAnterior, proceso.Estado)

	if destino.Estado == globals.EstadoReady {
		select {
		case globals.HayProcesoEnReady <- struct{}{}:
		default:
			// No hago nada si ya hay una señal pendiente (evito bloquear)
		}
	}

	/*if destino.Estado == globals.EstadoReady && globals.Config.PlanificacionCP == "SRT" {
		go VerificarDesalojoPorSRT(proceso)
	}*/
}

func BuscarPCBPorPID(pid int) *globals.PCB {
	colas := []*globals.Cola{
		globals.ColaNEW,
		globals.ColaREADY,
		globals.ColaEXEC,
		globals.ColaBLOCK,
		globals.ColaEXIT,
		globals.ColaSuspReady,
		globals.ColaSuspBlock,
	}

	for _, cola := range colas {
		cola.MutexCola.Lock()
		for _, pcb := range cola.Cola {
			if pcb.PID == pid {
				cola.MutexCola.Unlock()
				return pcb
			}
		}
		cola.MutexCola.Unlock()
	}

	return nil // No se encontró el proceso
}

func EliminarPCBPorPID(pid int) {
	colas := []*globals.Cola{
		globals.ColaNEW,
		globals.ColaREADY,
		globals.ColaEXEC,
		globals.ColaBLOCK,
		globals.ColaEXIT,
	}
	for _, cola := range colas {
		cola.MutexCola.Lock()
		nuevaCola := make([]*globals.PCB, 0, len(cola.Cola))
		for _, pcb := range cola.Cola {
			if pcb.PID != pid {
				nuevaCola = append(nuevaCola, pcb)
			}
		}
		cola.Cola = nuevaCola
		cola.MutexCola.Unlock()
	}

}

func SacarDeColaPorEstado(proceso *globals.PCB) {
	var cola *globals.Cola

	switch proceso.Estado {
	case globals.EstadoNew:
		cola = globals.ColaNEW
	case globals.EstadoReady:
		cola = globals.ColaREADY
	case globals.EstadoExec:
		cola = globals.ColaEXEC
	case globals.EstadoBlocked:
		cola = globals.ColaBLOCK
	case globals.EstadoSuspReady:
		cola = globals.ColaSuspReady
	case globals.EstadoSuspBlock:
		cola = globals.ColaSuspBlock
	case globals.EstadoExit:
		cola = globals.ColaEXIT
	default:
		log.Fatalf("Estado desconocido al sacar proceso: %s", proceso.Estado)
	}

	cola.MutexCola.Lock()
	defer cola.MutexCola.Unlock()

	for i, p := range cola.Cola {
		if p.PID == proceso.PID {
			cola.Cola = append(cola.Cola[:i], cola.Cola[i+1:]...)
			return
		}
	}
}

func ImprimirEstadoColas() {
	printCola := func(nombre string, cola *globals.Cola) {
		cola.MutexCola.Lock()
		defer cola.MutexCola.Unlock()

		if len(cola.Cola) == 0 {
			log.Printf("%s: [VACÍA]\n", nombre)
			return
		}

		log.Printf("%s: ", nombre)
		for _, p := range cola.Cola {
			log.Printf("%d ", p.PID)
		}
		log.Println()
	}

	log.Println("===== ESTADO DE LAS COLAS =====")
	printCola("NEW", globals.ColaNEW)
	printCola("READY", globals.ColaREADY)
	printCola("EXEC", globals.ColaEXEC)
	printCola("BLOCK", globals.ColaBLOCK)
	printCola("SUSP_READY", globals.ColaSuspReady)
	printCola("SUSP_BLOCK", globals.ColaSuspBlock)
	printCola("EXIT", globals.ColaEXIT)
	log.Println("================================")
}
