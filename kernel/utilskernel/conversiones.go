package utilskernel

import (
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func ConvertirAProcesoAInicializar(pcb *globals.PCB) mensajes.ProcesoAInicializar {
	return mensajes.ProcesoAInicializar{
		PID:           pcb.PID,
		NombreArchivo: pcb.Archivo,
		Tamanio:       pcb.Tamanio,
	}
}

func ConvertirAProcesoAEjecutar(pcb *globals.PCB) mensajes.ProcesoAEjecutar {
	return mensajes.ProcesoAEjecutar{
		PID: pcb.PID,
		PC:  pcb.PC,
	}
}
