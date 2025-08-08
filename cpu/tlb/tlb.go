package tlb

import (
	"sync"
	"time"
	"log"
)

type TLBEntry struct {
	PID       int
	NroPagina int
	Marco     int
	Timestamp int64 //PAra el lru
	Order     int   //PAra el fifo
}

//Nota Cuando Desalojo un proceso debo eliminar todas las paginas de la TLB pertenecientes

var (
	tlb                   []TLBEntry
	tamanioMaximoEntradas int
	algoritmo             string
	ordenFIFO             int
	mutex                 sync.Mutex
)

// Creo la funcion para configurar la TLB
func Init(cantidadEntradas int, algoritmoReemplazo string) {
	mutex.Lock()
	defer mutex.Unlock()

	if algoritmoReemplazo != "FIFO" && algoritmoReemplazo != "LRU" {
		log.Fatalf("Algoritmo de reemplazo inválido: %s", algoritmoReemplazo)
	}

	tamanioMaximoEntradas = cantidadEntradas
	algoritmo = algoritmoReemplazo
	tlb = make([]TLBEntry, 0, cantidadEntradas)
	ordenFIFO = 0
}

// Busco la pagina en la tlb y doy el marco (TLB HIT)
func TraducirPagina(pid int, nroPagina int) (marco int, tlbHit bool) {
	mutex.Lock()
	defer mutex.Unlock()

	for i, entrada := range tlb {
		if entrada.PID == pid && entrada.NroPagina == nroPagina {
			tlb[i].Timestamp = time.Now().UnixNano() // para LRU

			return entrada.Marco, true
		}
	}
	return -1, false

}

func AgregarEntrada(pid int, nroPagina int, marco int) {
	//Chequeo si la tlb esta desabilitada

	if tamanioMaximoEntradas == 0 {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	nueva := TLBEntry{
		PID:       pid,
		NroPagina: nroPagina,
		Marco:     marco,
		Timestamp: time.Now().UnixNano(),
		Order:     ordenFIFO,
	}
	ordenFIFO++

	if len(tlb) < tamanioMaximoEntradas {
		tlb = append(tlb, nueva)
	} else {
		switch algoritmo {
		case "LRU":
			AlgortimoDeReemplazoLRU(nueva)
		case "FIFO":
			AlgortimoDeReemplazoFIFO(nueva)
		}
	}
}

// algoritmos que seleccionan victimas
func AlgortimoDeReemplazoLRU(entradaNueva TLBEntry) {
	iDeReemplazo := 0
	menorTiempo := tlb[0].Timestamp

	for i, entrada := range tlb { //recorre la tlb
		if entrada.Timestamp < menorTiempo {
			menorTiempo = entrada.Timestamp
			iDeReemplazo = i // va actualizando cual es la posible victima
		}
	}
	tlb[iDeReemplazo] = entradaNueva //reemplazo con la nueva entrada
}

func AlgortimoDeReemplazoFIFO(nuevaEntrada TLBEntry) {
	indiceReemplazo := 0
	menorOrden := tlb[0].Order

	for i, entrada := range tlb {
		if entrada.Order < menorOrden {
			menorOrden = entrada.Order
			indiceReemplazo = i
		}
	}
	tlb[indiceReemplazo] = nuevaEntrada
}

//Vacio tlb cuando el proceso es desalojado

func LimpiarTLB(pid int) {
	mutex.Lock()
	defer mutex.Unlock()

	nuevaTLB := make([]TLBEntry, 0)
	for _, entrada := range tlb {
		if entrada.PID != pid {
			nuevaTLB = append(nuevaTLB, entrada)
		}
	}
	tlb = nuevaTLB
}
