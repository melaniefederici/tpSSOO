package utilsmemoria

import (
	"log"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func SuspenderProceso(pid int) {
	GuardarPaginasEnSwap(pid)
	LiberarMarcos(pid)
	delete(globals.TablasPorProceso, pid)

	globals.MetricasPorProceso[pid].BajadasASwap++
	log.Printf("Se suspendio a swap el proceso con el PID: %d", pid)
}

func DesuspenderProceso(pid int) {
	proceso := globals.ProcesosCargados[pid]

	// cuando se usa PCMP, el proceso con PID 5 tiene un tamaño de 256 (todo el espacio en memoria)
	// justo despues el proceso PID 1 intenta Desuspenderse, pero falla porque cuando se intentea reservar la memoria no deja.
	// creo que habria que hacer un chequeo antes tipo "HayEspacioDisponible"
	ReservarMemoria(pid, proceso.Tamanio)

	CargarPaginasDesdeSwap(pid)
	LiberarSwap(pid)

	globals.MetricasPorProceso[pid].SubidasAMemoriaPrincipal++
	log.Printf("Se restauro de swap el proceso con el PID: %d", pid)
}

func GuardarPaginasEnSwap(pid int) {
	globals.SwapFileMutex.Lock()
	defer globals.SwapFileMutex.Unlock()

	//aca abro o creo swapfile
	swapFile, err := os.OpenFile(globals.Config.SwapfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("No se pudo abrir el swapfile: %v", err)
	}
	defer swapFile.Close()

	//busco la tabla con el pid
	tabla, ok := globals.TablasPorProceso[pid]
	if !ok {
		log.Printf("No se encontró tabla de páginas del proceso con el PID: %d", pid)
		return
	}

	offset, _ := swapFile.Seek(0, os.SEEK_END)

	numPagina := 0
	tamanio := globals.Config.PageSize

	RecorrerTabla(tabla.Tabla, func(entrada *globals.Entrada) {
		marco := entrada.Final.Marco
		contenido := globals.MemoriaPrincipal[marco*tamanio : (marco+1)*tamanio]

		n, err := swapFile.Write(contenido)
		if err != nil || n != tamanio {
			log.Fatalf("Error al escribir contenido en swapfile.bin: %v", err)
		}

		globals.TablaSwap = append(globals.TablaSwap, globals.EntradaSwap{
			PID:     pid,
			Pagina:  numPagina,
			Offset:  int(offset),
			Tamanio: tamanio,
		})

		log.Printf("# PID: %d, la página %d fue guardada en SWAP", pid, numPagina)

		offset += int64(tamanio)
		numPagina++
	})

	log.Printf("El proceso con el PID %d fue guardado en SWAP", tabla.PID)
}

func CargarPaginasDesdeSwap(pid int) {
	//abro swapfile
	swapFile, err := os.Open(globals.Config.SwapfilePath)
	if err != nil {
		log.Fatalf("Error al abrir swapfile: %v", err)
	}
	defer swapFile.Close()

	tamanio := globals.Config.PageSize

	//busco paginas del procesp
	for _, entrada := range globals.TablaSwap {
		if entrada.PID != pid {
			continue
		}

		//leo pagina
		contenido := make([]byte, entrada.Tamanio)
		_, err := swapFile.ReadAt(contenido, int64(entrada.Offset))
		if err != nil {
			log.Fatalf("Error al leer página desde swap: %v", err)
		}

		//busco marco y escribo en memoria ppal
		marco := ObtenerMarcoSwap(pid, entrada.Pagina)
		copy(globals.MemoriaPrincipal[marco*tamanio:(marco+1)*tamanio], contenido)
	}
	log.Printf("El proceso con el PID %d fue restaurado desde swap", pid)
}

func ObtenerMarcoSwap(pid int, nroPagina int) int {
	entries := globals.Config.EntriesPerPage
	niveles := globals.Config.NumberOfLevels

	entradaActual := globals.TablasPorProceso[pid].Tabla

	for nivel := 1; nivel < niveles; nivel++ {
		indice := (nroPagina / pow(entries, niveles-nivel)) % entries
		if indice >= len(entradaActual) || entradaActual[indice] == nil {
			log.Fatalf("Entrada no válida en nivel %d (índice %d)", nivel, indice)
		}
		entradaActual = entradaActual[indice].Nivel
	}

	// Llegamos al último nivel
	indiceFinal := (nroPagina / pow(entries, 0)) % entries
	if indiceFinal >= len(entradaActual) || entradaActual[indiceFinal] == nil || entradaActual[indiceFinal].Final == nil {
		log.Fatalf("Entrada final no válida para página %d", nroPagina)
	}

	return entradaActual[indiceFinal].Final.Marco

}

func LiberarSwap(pid int) {
	globals.SwapFileMutex.Lock()
	defer globals.SwapFileMutex.Unlock()

	tablaSwap := make([]globals.EntradaSwap, 0)

	for _, entrada := range globals.TablaSwap {
		if entrada.PID != pid {
			tablaSwap = append(tablaSwap, entrada)
		}
	}

	globals.TablaSwap = tablaSwap
	log.Printf("Se liberó el espacio en el swapfile del proceso con el PID: %d", pid)
}
