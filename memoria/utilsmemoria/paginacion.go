package utilsmemoria

import (
	"log"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func CantidadDePaginasPorProceso(tamanio int) int {
	pageSize := globals.Config.PageSize
	paginas := tamanio / pageSize
	if tamanio%pageSize != 0 {
		paginas++ //si sobra algo, agrego una pagina final
	}
	return paginas
}

func CrearTabla(nivelActual, nivelMaximo int, paginasRestantes *int) []*globals.Entrada {
	entries := globals.Config.EntriesPerPage
	tabla := make([]*globals.Entrada, 0, entries)

	for i := 0; i < entries && *paginasRestantes > 0; i++ {
		entrada := &globals.Entrada{}

		if nivelActual == nivelMaximo {
			// Hoja: asignar marco
			marco := BuscarMarcoLibre()
			if marco == -1 {
				log.Fatal("No hay marcos libres")
			}
			entrada.Final = &globals.EntradaFinal{Marco: marco}
			*paginasRestantes--
		} else {
			// Nivel intermedio: seguir recursivamente
			entrada.Nivel = CrearTabla(nivelActual+1, nivelMaximo, paginasRestantes)
		}

		tabla = append(tabla, entrada)
	}

	return tabla
}

func RecorrerTabla(tabla []*globals.Entrada, accion func(entrada *globals.Entrada)) {
	for _, entrada := range tabla {
		if entrada.Final != nil {
			accion(entrada)
		} else if entrada.Nivel != nil {
			RecorrerTabla(entrada.Nivel, accion)
		}
	}
}

func BuscarMarcoLibre() int {
	//recorro el bitmap, si hay un marco libre lo ocupo
	for i, libre := range globals.MarcosLibres {
		if libre {
			globals.MarcosLibres[i] = false
			return i
		}
	}
	return -1
}

func contarMarcosLibres() int {
	libres := 0
	for _, libre := range globals.MarcosLibres {
		if libre {
			libres++
		}
	}
	return libres
}

func InicializarMarcos() {
	//inicializo estructuras
	cantidadMarcos := globals.Config.MemorySize / globals.Config.PageSize
	globals.MarcosLibres = make([]bool, cantidadMarcos)
	for i := range globals.MarcosLibres {
		globals.MarcosLibres[i] = true
	}
}

func LiberarMarcos(pid int) {
	tabla, ok := globals.TablasPorProceso[pid]
	if !ok {
		log.Printf("No se encontro la tabla del proceso %d", pid)
		return
	}

	RecorrerTabla(tabla.Tabla, func(entrada *globals.Entrada) {
		globals.MarcosLibres[entrada.Final.Marco] = true
		log.Printf("Liberado el marco %d", entrada.Final.Marco)
	})
}

// Parte correspondiente al acceso de tabla de paginas

//esto responde a cpu (tlb)

func ObtenerMarco(pid int, nroPagina int) int {
	delay := time.Duration(globals.Config.MemoryDelay) * time.Millisecond
	entries := globals.Config.EntriesPerPage
	niveles := globals.Config.NumberOfLevels

	metricas := globals.MetricasPorProceso[pid]

	indices := CalcularIndices(nroPagina, niveles, entries)

	// por cada nivel simulo delay
	for i := 0; i < niveles; i++ {
		time.Sleep(delay)
		metricas.AccesosATabla++
	}

	entrada := BuscarEntrada(globals.TablasPorProceso[pid].Tabla, indices)

	log.Printf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, nroPagina, entrada.Final.Marco)
	return entrada.Final.Marco
}

func CalcularIndices(pagina int, niveles int, entries int) []int {
	indices := make([]int, niveles)
	for i := 0; i < niveles; i++ {
		exp := niveles - i - 1
		indices[i] = (pagina / pow(entries, exp)) % entries
	}
	return indices
}

func BuscarEntrada(tabla []*globals.Entrada, indices []int) *globals.Entrada {
	actual := tabla
	for _, idx := range indices {
		if idx >= len(actual) || actual[idx] == nil {
			return nil
		}
		if actual[idx].Final != nil {
			return actual[idx]
		}
		actual = actual[idx].Nivel
	}
	return nil
}

func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
