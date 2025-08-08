package utilsmemoria

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func RespuestaAIniciarProceso(w http.ResponseWriter, r *http.Request) {

	var proceso mensajes.ProcesoAInicializar
	if err := json.NewDecoder(r.Body).Decode(&proceso); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	rutaBase := "../revenge-of-the-cth-pruebas/"
	rutaArchivo := filepath.Join(rutaBase, proceso.NombreArchivo)

	if HayEspacioDisponible(proceso.Tamanio) {
		ReservarMemoria(proceso.PID, proceso.Tamanio)
		instrucciones := CargarInstruccionesDesdeArchivo(rutaArchivo)
		if instrucciones == nil {
			http.Error(w, "No se pudo leer el archivo", http.StatusInternalServerError)
			return
		}

		globals.ProcesosCargados[proceso.PID] = globals.ProcesoEnMemoria{
			PID:           proceso.PID,
			Instrucciones: instrucciones,
			Tamanio:       proceso.Tamanio,
		}

		globals.MetricasPorProceso[proceso.PID] = &globals.Metricas{} //inicializo metricas del proceso

		log.Printf("Cargadas %d instrucciones para PID %d desde %s", len(instrucciones), proceso.PID, proceso.NombreArchivo)
		log.Printf("## PID: %d - Proceso Creado - Tamanio: %d", proceso.PID, proceso.Tamanio)

		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	} else {
		json.NewEncoder(w).Encode(map[string]bool{"ok": false})
	}
}

func RespuestaAFinalizarProceso(w http.ResponseWriter, r *http.Request) {
	var datos struct {
		PID int `json:"pid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		log.Printf("Error al decodificar solicitud de finalización: %s", err.Error())
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	log.Printf("Memoria recibió solicitud para finalizar el proceso con PID %d", datos.PID)

	ok := FinalizarProceso(datos.PID)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]bool{"ok": false})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func RespuestaDumpMemory(w http.ResponseWriter, r *http.Request) {
	var datos struct {
		PID int `json:"pid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	DumpDeProceso(datos.PID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func RespuestaAPedidoSusp(w http.ResponseWriter, r *http.Request) {
	var pedido struct {
		PID int `json:"pid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		log.Printf("Error al decodificar el pedido de suspension del proceso con el PID: %d", pedido.PID)
		http.Error(w, "Solicitud invalida", http.StatusBadRequest)
		return
	}

	SuspenderProceso(pedido.PID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func RespuestaAPedidoUnsusp(w http.ResponseWriter, r *http.Request) {
	var pedido struct {
		PID int `json:"pid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		log.Printf("Error al decodificar el pedido de desuspension del proceso con el PID: %d", pedido.PID)
		http.Error(w, "Solicitud invalida", http.StatusBadRequest)
		return
	}

	tamanio := globals.TamaniosPorProceso[pedido.PID]

	if !HayEspacioDisponible(tamanio) {
		log.Printf("No hay espacio para des-suspender al proceso con PID %d", pedido.PID)
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("NO_HAY_ESPACIO"))
		return
	}

	DesuspenderProceso(pedido.PID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
