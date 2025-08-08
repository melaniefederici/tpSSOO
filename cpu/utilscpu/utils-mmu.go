package utilscpu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

func PaginaPresente(pid int, nroPagina int) bool {
	paginas, ok := globals.Bitmap[pid]
	if !ok || nroPagina >= len(paginas) {
		return false
	}
	return paginas[nroPagina]
}

func CargarPagina(pid int, nroPagina int, marco int) {
	globals.Mutex.Lock()
	defer globals.Mutex.Unlock()

	extenderEstructuras(pid, nroPagina)

	globals.Bitmap[pid][nroPagina] = true
	globals.TablaMarcos[pid][nroPagina] = marco
}

func extenderEstructuras(pid int, hastaPagina int) {
	for len(globals.Bitmap[pid]) <= hastaPagina {
		globals.Bitmap[pid] = append(globals.Bitmap[pid], false)
		globals.TablaMarcos[pid] = append(globals.TablaMarcos[pid], -1)
	}
}

func ObtenerMarcoDesdeMemoria(pid int, nroPagina int) (int, error) {
	url := fmt.Sprintf("http://%s:%d/marco", globals.Config.IPMemoria, globals.Config.PuertoMemoria)

	req := map[string]int{
		"pid":       pid,
		"nroPagina": nroPagina,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("error serializando petición de marco: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("error consultando marco a memoria: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("respuesta no OK: %s", resp.Status)
	}

	var respuesta struct {
		Marco int `json:"marco"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respuesta)
	if err != nil {
		return 0, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	return respuesta.Marco, nil
}

func LimpiarBitmapYTablas(pid int) {
	globals.Mutex.Lock()
	defer globals.Mutex.Unlock()

	delete(globals.TablaMarcos, pid)
	delete(globals.Bitmap, pid)
}
