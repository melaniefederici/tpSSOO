package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

type EntradaCache struct {
	PID        int
	NroPagina  int
	Contenido  []byte
	Modificada bool
	Usada      bool //para CLOCK
}

var (
	cache              []EntradaCache
	cacheMaxEntries    int
	algoritmoReemplazo string
	clockPointer       int
	cacheMutex         sync.Mutex
	delayMilisegundos  int
)

//inicializo la cache

func Init(entries int, algoritmo string, delay int) {
	cacheMaxEntries = entries
	algoritmoReemplazo = algoritmo
	delayMilisegundos = delay
	cache = make([]EntradaCache, 0, entries)
	clockPointer = 0
}

// Leo la pagina
func LeerPagina(pid int, nroPagina int) ([]byte, bool) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	for i, entrada := range cache {
		if entrada.PID == pid && entrada.NroPagina == nroPagina {
			cache[i].Usada = true
			time.Sleep(time.Duration(delayMilisegundos) * time.Millisecond)
			log.Printf("PID: %d - Cache Hit - Pagina: %d", pid, nroPagina)
			log.Printf("El contenido es: %v", entrada.Contenido)
			return entrada.Contenido, true
		}
	}

	log.Printf("PID: %d - Cache Miss - Pagina: %d", pid, nroPagina)
	return nil, false
}

//Escribo pagina

func EscribirPagina(pid int, nroPagina int, contenido []byte) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	for i, entrada := range cache {
		if entrada.PID == pid && entrada.NroPagina == nroPagina {
			cache[i].Contenido = contenido
			cache[i].Modificada = true
			cache[i].Usada = true
			time.Sleep(time.Duration(delayMilisegundos) * time.Millisecond)
			//log.Printf("PID: %d - Cache Hit - Pagina: %d (escritura)", pid, nroPagina)
			return
		}
	}

	//log.Printf("PID: %d - Cache Miss - Pagina: %d (escritura)", pid, nroPagina)
	AgregarEntrada(pid, nroPagina, contenido, true)
}

//Funcion de agregar entrada

func AgregarEntrada(pid int, nroPagina int, contenido []byte, modificada bool) {
	if cacheMaxEntries == 0 {
		return // Cache deshabilitada
	}

	// Ya existe la entrada, no agrego duplicado
	for _, entrada := range cache {
		if entrada.PID == pid && entrada.NroPagina == nroPagina {
			return
		}
	}

	nueva := EntradaCache{
		PID:        pid,
		NroPagina:  nroPagina,
		Contenido:  contenido,
		Modificada: modificada,
		Usada:      true,
	}

	if len(cache) < cacheMaxEntries {
		cache = append(cache, nueva)
		log.Printf("PID: %d - Cache Add - Pagina: %d", pid, nroPagina)
		return
	}

	if algoritmoReemplazo == "CLOCK" {
		reemplazoCLOCK(nueva)
	} else {
		reemplazoCLOCKM(nueva)
	}
}

func reemplazoCLOCK(nueva EntradaCache) {
	for {
		if !cache[clockPointer].Usada {
			log.Printf("PID: %d - Reemplazo CLOCK - Pagina: %d", cache[clockPointer].PID, cache[clockPointer].NroPagina)

			if cache[clockPointer].Modificada {
				escribirAMemoria(cache[clockPointer]) // función a definir
			}

			cache[clockPointer] = nueva
			log.Printf("PID: %d - Cache Add - Pagina: %d", nueva.PID, nueva.NroPagina)

			clockPointer = (clockPointer + 1) % cacheMaxEntries
			break
		}

		cache[clockPointer].Usada = false
		clockPointer = (clockPointer + 1) % cacheMaxEntries
	}
}

func reemplazoCLOCKM(nueva EntradaCache) {
	pasos := 0
	inicio := clockPointer

	for {

		//Primera vuelta
		if pasos == 0 {
			log.Println("CLOCK-M Paso 1: Buscando entrada con U=0 y M=0")
			mostrarEstadoCache()
			inicio = clockPointer
			//Recorro la cache y  busco (0,0)
			for i := 0; i < cacheMaxEntries; i++ {
				entrada := &cache[clockPointer]

				if !entrada.Usada && !entrada.Modificada {
					log.Printf("CLOCK-M Paso 1: Reemplazo directo - Página víctima: %d", entrada.NroPagina)
					cache[clockPointer] = nueva
					log.Printf("PID: %d - Cache Add - Página: %d", nueva.PID, nueva.NroPagina)
					avanzarClock()
					return
				}
				avanzarClock()
			}
			//si no lo encontre entonces avanzo a la segunda vuelta
			pasos++
		}

		//segunda vuelta
		if pasos == 1 {
			log.Println("CLOCK-M Paso 2: Buscando entrada con U=0 y M=1")
			mostrarEstadoCache()
			inicio = clockPointer
			//Busco (0,1), si no esta (0,1) voy modificando el bit de uso en cero
			for i := 0; i < cacheMaxEntries; i++ {
				entrada := &cache[clockPointer]

				if !entrada.Usada && entrada.Modificada {
					log.Printf("CLOCK-M Paso 2: Entrada modificada U=0 M=1, se escribe antes de reemplazar - Página víctima: %d", entrada.NroPagina)
					escribirAMemoria(*entrada)
					cache[clockPointer] = nueva
					log.Printf("PID: %d - Cache Add - Página: %d", nueva.PID, nueva.NroPagina)
					avanzarClock()
					return
				}

				if entrada.Usada {
					entrada.Usada = false
				}
				avanzarClock()
			}
		}

		log.Println("CLOCK-M: No se encontró víctima, reiniciando búsqueda desde paso 1")
		pasos = 0
		clockPointer = inicio //reinicio puntero al inicio de la primera vuelta
	}
}

func avanzarClock() {
	clockPointer = (clockPointer + 1) % cacheMaxEntries
	log.Printf("CLOCK-M: Puntero avanzado a posición: %d", clockPointer)
}

func mostrarEstadoCache() {
	log.Println("📦 Estado actual de la CACHE:")
	for i, entrada := range cache {
		log.Printf("[%d] PID: %d | Página: %d | Usada: %v | Modificada: %v", i, entrada.PID, entrada.NroPagina, entrada.Usada, entrada.Modificada)
	}
	log.Printf("Puntero CLOCK en posición: %d", clockPointer)
	log.Println("-----------------------------")
}

// vacio la cache cuando el proceso es desalojado y mando a memoria principal las paginas modificadas

func LimpiarProceso(pid int) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	nuevaCache := make([]EntradaCache, 0)

	for _, entrada := range cache {
		if entrada.PID == pid {
			if entrada.Modificada {
				escribirAMemoria(entrada) // función a definir
			}
		} else {
			nuevaCache = append(nuevaCache, entrada)
		}
	}

	cache = nuevaCache
}

// funcion para comunicarme con memoria escribirAMemoria revisar al desarrollar el modulo de memoria

func escribirAMemoria(entrada EntradaCache) {
	url := fmt.Sprintf("http://%s:%d/solicitudActualizarPagina", globals.Config.IPMemoria, globals.Config.PuertoMemoria)

	tamPagina := globals.Config.TamanioPagina
	dirFisica := entrada.NroPagina * tamPagina

	body := map[string]interface{}{
		"pid":        entrada.PID,
		"dir_fisica": dirFisica,
		"contenido":  string(entrada.Contenido),
	}

	//HACER QUE MATCHEE CON RespuestaActualizarPagina o algo asi

	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("ERROR escribiendo página modificada en memoria: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("PID: %d - Memory Update - Página: %d", entrada.PID, entrada.NroPagina)
	} else {
		log.Printf("Error al escribir página en memoria (HTTP %d)", resp.StatusCode)
	}
}
