package utilsmemoria

import (
	"log"
	"os"
	"strings"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func CargarInstruccionesDesdeArchivo(nombreArchivo string) []string {
	contenido, err := os.ReadFile(nombreArchivo)
	if err != nil {
		return nil
	}

	lineas := strings.Split(string(contenido), "\n")
	var instrucciones []string
	for _, linea := range lineas {
		trimmed := strings.TrimSpace(linea)
		if trimmed != "" {
			instrucciones = append(instrucciones, trimmed)
		}
	}
	return instrucciones
}

func BuscarInstruccion(pid, pc int) (string, []string) {
	proceso, ok := globals.ProcesosCargados[pid]

	if !ok {
		log.Printf("No se encontro al proceso con el pid: %d", pid)
		return "", nil
	}

	if pc < 0 || pc > len(proceso.Instrucciones) {
		log.Printf("El Program Counter %d no es valido", pc)
		return "", nil
	}

	instruccionCompleta := proceso.Instrucciones[pc]
	instruccionPorPartes := strings.Fields(instruccionCompleta)

	instruccion := instruccionPorPartes[0]
	parametros := instruccionPorPartes[1:]

	metrica := globals.MetricasPorProceso[pid]
	metrica.InstruccionesSolicitadas++

	return instruccion, parametros

}
