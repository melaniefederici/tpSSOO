package utilskernel

import (
	"log"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
)

func generarNuevoPID() int {
	pid := globals.PIDActual
	globals.PIDActual++
	return pid
}

func CrearPCB(nombreArchivo string, tamanio int) {
	pid := generarNuevoPID()

	me := map[globals.EstadoProceso]int{
		globals.EstadoNew:     1,
		globals.EstadoReady:   0,
		globals.EstadoExec:    0,
		globals.EstadoBlocked: 0,
		globals.EstadoExit:    0,
	}

	mt := map[globals.EstadoProceso]int{
		globals.EstadoNew:     0,
		globals.EstadoReady:   0,
		globals.EstadoExec:    0,
		globals.EstadoBlocked: 0,
		globals.EstadoExit:    0,
	}

	pcb := &globals.PCB{
		PID:              pid,
		PC:               0,
		Estado:           globals.EstadoNew,
		Archivo:          nombreArchivo,
		Tamanio:          tamanio,
		ME:               me,
		MT:               mt,
		EstimacionRafaga: globals.Config.EstimacionInicial, //cheaquear esto
		RafagaReal:       0,
		InicioEstado:     time.Now(),
	}

	log.Printf("## (%d) Se crea el proceso - Estado: %s", pcb.PID, pcb.Estado)
	globals.ColaNEW.MutexCola.Lock()
	globals.ColaNEW.Cola = append(globals.ColaNEW.Cola, pcb)
	globals.ColaNEW.MutexCola.Unlock()
}
