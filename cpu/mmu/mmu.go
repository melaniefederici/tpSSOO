package mmu

import (
	"log"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/tlb"
	"github.com/sisoputnfrba/tp-golang/cpu/utilscpu"
)

func TraducirDireccion(dirLogica int, pid int) int {
	tamPagina := globals.Config.TamanioPagina
	nroPagina := dirLogica / tamPagina
	desplazamiento := dirLogica % tamPagina

	//que busque primero en la tlb
	marco, hit := tlb.TraducirPagina(pid, nroPagina)

	if !hit {
		log.Printf("PID: %d - TLB MISS - Pagina: %d", pid, nroPagina)

		globals.Mutex.Lock()
		if !utilscpu.PaginaPresente(pid, nroPagina) {
			globals.Mutex.Unlock()
			log.Printf("PID: %d - Página %d no está presente (bitmap = false)", pid, nroPagina)
			marcoMemoria, err := utilscpu.ObtenerMarcoDesdeMemoria(pid, nroPagina)
			if err != nil {
				log.Printf("Error consultando marco a memoria: %v", err)
				return -1
			}

			utilscpu.CargarPagina(pid, nroPagina, marcoMemoria)

			// actualizo tlb
			tlb.AgregarEntrada(pid, nroPagina, marcoMemoria)
			marco = marcoMemoria

		} else {
			//Pagina en tabla pero no en tlb
			marco = globals.TablaMarcos[pid][nroPagina]

			globals.Mutex.Unlock()

			tlb.AgregarEntrada(pid, nroPagina, marco)
		}
	} else {
		log.Printf("PID: %d - TLB HIT - Pagina: %d", pid, nroPagina)
	}

	// hay que cargar en cache si es que estaba en tlb o en tabla
	log.Printf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, nroPagina, marco)
	direccionFisica := marco*tamPagina + desplazamiento

	log.Printf("PID: %d - Dirección lógica: %d -> Dirección física: %d", pid, dirLogica, direccionFisica)
	return direccionFisica
}
