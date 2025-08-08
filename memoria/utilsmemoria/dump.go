package utilsmemoria

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func DumpDeProceso(pid int) {
	proceso, ok := globals.ProcesosCargados[pid]
	if !ok {
		log.Printf("No se encontró el proceso con PID %d", pid)
		return
	}

	tamanio := proceso.Tamanio
	memoriaCompleta := globals.MemoriaPrincipal

	// Obtener timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Crear nombre del archivo
	filename := fmt.Sprintf("%d-%s.dmp", pid, timestamp)

	// Ruta final
	fullPath := filepath.Join(globals.Config.DumpPath, filename)

	// Crear carpeta y archivo
	os.MkdirAll(globals.Config.DumpPath, 0755)
	file, err := os.Create(fullPath)
	if err != nil {
		log.Printf("Error creando archivo de dump: %v", err)
		return
	}
	defer file.Close()

	// Escribir contenido de memoria correspondiente a ese proceso
	// Suponemos que las páginas están asignadas a marcos físicos
	// y que su contenido está contiguo en memoria.
	for i := 0; i < tamanio; i++ {
		file.Write([]byte{memoriaCompleta[i]})
	}

	log.Printf("## PID: %d - Memory Dump solicitado", pid)
	log.Printf("Dump guardado en: %s", fullPath)
}
