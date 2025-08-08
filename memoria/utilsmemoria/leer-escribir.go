package utilsmemoria

import "github.com/sisoputnfrba/tp-golang/memoria/globals"

func LeerContenido(pid, dirFisica int, tamanio int) []byte {
	var contenido []byte
	for i := 0; i < tamanio; i++ {
		contenido = append(contenido, globals.MemoriaPrincipal[dirFisica+i])
	}
	globals.MetricasPorProceso[pid].LecturasDeMemoria++
	return contenido

}

func EscribirContenido(pid, dirFisica int, cadena string) {
	bytes := []byte(cadena)
	for i, b := range bytes {
		globals.MemoriaPrincipal[dirFisica+i] = b
	}
	globals.MetricasPorProceso[pid].EscriturasDeMemoria++
}

func LeerPaginaCompleta(pid, dirFisica int) []byte {
	tamanio := globals.Config.PageSize
	// Calcular la dirección de inicio de página (alinear hacia abajo)
	dirPagina := (dirFisica / tamanio) * tamanio
	globals.MetricasPorProceso[pid].LecturasDeMemoria++
	return globals.MemoriaPrincipal[dirPagina : dirPagina+tamanio]
}

func ActualizarPaginaCompleta(pid, dirFisica int, cadena string) {
	bytes := []byte(cadena)
	tamanio := globals.Config.PageSize
	globals.MetricasPorProceso[pid].EscriturasDeMemoria++
	for i := 0; i < tamanio; i++ {
		globals.MemoriaPrincipal[dirFisica+i] = bytes[i]
	}
}
