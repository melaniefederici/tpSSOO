package main

import (
	"net/http"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/memoria/utilsmemoria"
	"github.com/sisoputnfrba/tp-golang/utils/inicio"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func main() {
	inicio.ConfigurarLogger("memoria.log")
	globals.Config = inicio.CargarConfiguracion[globals.MemoriaConfig]("memoria-config.json")

	globals.MemoriaPrincipal = make([]byte, globals.Config.MemorySize)
	utilsmemoria.InicializarMarcos()

	handlers := map[string]http.HandlerFunc{
		"/mensaje":                      mensajes.RecibirMensaje,
		"/paquetes":                     mensajes.RecibirPaquetes,
		"/procesoAInicializar":          mensajes.RecibirProcesoAInicializar,
		"/solicitudProcesoAInicializar": utilsmemoria.RespuestaAIniciarProceso,
		"/solicitudFinalizarProceso":    utilsmemoria.RespuestaAFinalizarProceso,
		"/solicitudInstruccion":         utilsmemoria.RespuestaInstruccion,
		"/solicitudEscritura":           utilsmemoria.RespuestaEscritura,
		"/solicitudLectura":             utilsmemoria.RespuestaLectura,
		"/solicitudActualizarPagina":    utilsmemoria.RespuestaActualizarPagina,
		"/solicitudLecturaPagina":       utilsmemoria.RespuestaLecturaPagina,
		"/solicitudSwapDeProceso":       utilsmemoria.RespuestaAPedidoSusp,
		"/solicitudDesuspender":         utilsmemoria.RespuestaAPedidoUnsusp,
		"/marco":                        utilsmemoria.RespuestaObtenerMarco,
		"/solicitudDumpMemory":          utilsmemoria.RespuestaDumpMemory,
		"/solicitudConfig":              utilsmemoria.ObtenerConfigHandler,
	}

	//levanto el servidor, por cada solicitud hace un hilo
	inicio.IniciarServidor(globals.Config.PuertoMemoria, handlers)
}
