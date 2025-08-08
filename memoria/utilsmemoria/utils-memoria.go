package utilsmemoria

import (
	"log"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func ReservarMemoria(pid int, tamanio int) {
	cantPaginas := CantidadDePaginasPorProceso(tamanio)
	tabla := CrearTabla(1, globals.Config.NumberOfLevels, &cantPaginas)

	//guardo la estructura con el pid como key
	globals.TablasPorProceso[pid] = &globals.TablaDePaginas{
		PID:     pid,
		Tabla:   tabla,
		Tamanio: tamanio,
	}

	globals.TamaniosPorProceso[pid] = tamanio
}

func HayEspacioDisponible(tamanioNecesario int) bool {
	paginas := CantidadDePaginasPorProceso(tamanioNecesario)
	marcosLibres := contarMarcosLibres()
	if paginas <= marcosLibres {
		return true
	}
	log.Printf("No hay suficientes marcos libres: necesito %d, disponibles %d", paginas, marcosLibres)
	return false
}

func FinalizarProceso(pid int) bool {

	LiberarMarcos(pid)

	delete(globals.TablasPorProceso, pid)
	delete(globals.ProcesosCargados, pid)

	metrica, ok := globals.MetricasPorProceso[pid]
	if !ok {
		metrica = &globals.Metricas{}
	}
	log.Printf("## PID: %d - Proceso Destruido - Metricas - Acc.T.Pag: %d; Inst.Sol: %d; SWAP: %d; Mem.Prin: %d; Lec.Mem: %d; Esc.Mem: %d",
		pid, metrica.AccesosATabla, metrica.InstruccionesSolicitadas, metrica.BajadasASwap, metrica.SubidasAMemoriaPrincipal, metrica.LecturasDeMemoria, metrica.EscriturasDeMemoria)

	delete(globals.MetricasPorProceso, pid)
	return true
}
