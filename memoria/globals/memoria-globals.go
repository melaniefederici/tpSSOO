package globals

import "sync"

type MemoriaConfig struct {
	PuertoMemoria  int    `json:"puerto_memoria"`
	IPMemory       string `json:"ip_memory"`
	MemorySize     int    `json:"memory_size"`
	PageSize       int    `json:"page_size"`
	EntriesPerPage int    `json:"entries_per_page"`
	NumberOfLevels int    `json:"number_of_levels"`
	MemoryDelay    int    `json:"memory_delay"`
	SwapfilePath   string `json:"swapfile_path"`
	SwapDelay      int    `json:"swap_delay"`
	LogLevel       string `json:"log_level"`
	DumpPath       string `json:"dump_path"`
	ScriptsPath    string `json:"strings_path"`
}

var Config *MemoriaConfig

var MemoriaPrincipal []byte

type ProcesoEnMemoria struct {
	PID           int
	Instrucciones []string
	Tamanio       int
}

var ProcesosCargados = make(map[int]ProcesoEnMemoria) // mapeo el PID con el Proceso

type PeticionInstruccion struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type RespuestaInstruccion struct {
	Instruccion string
	Parametros  []string
}

type PeticionLectura struct {
	PID       int `json:"pid"`
	DirFisica int `json:"dir_fisica"`
	Tamanio   int `json:"tamanio"`
}

type RespuestaLectura struct {
	Valor byte `json:"valor"`
}

type PeticionEscritura struct {
	PID       int    `json:"pid"`
	DirFisica int    `json:"dir_fisica"`
	Cadena    string `json:"valor"`
}

type PeticionLecturaPagina struct {
	PID       int `json:"pid"`
	DirFisica int `json:"dir_fisica"`
}

type RespuestaLecturaPagina struct {
	Contenido []byte `json:"contenido"`
}

type PeticionActualizarPagina struct {
	PID       int    `json:"pid"`
	DirFisica int    `json:"dir_fisica"`
	Cadena    string `json:"contenido"`
}

// Estructuras necesarias para paginacion jerarquica multinivel
type Entrada struct {
	Nivel []*Entrada
	Final *EntradaFinal
}

type EntradaFinal struct {
	Marco int
}

type TablaDePaginas struct {
	PID     int
	Tabla   []*Entrada
	Tamanio int
}

var TablasPorProceso = make(map[int]*TablaDePaginas)
var MarcosLibres []bool // Bitmap TRUE: Libre, FALSE: Ocupado

type Metricas struct {
	AccesosATabla            int
	InstruccionesSolicitadas int
	BajadasASwap             int
	SubidasAMemoriaPrincipal int
	LecturasDeMemoria        int
	EscriturasDeMemoria      int
}

var MetricasPorProceso = make(map[int]*Metricas)

type EntradaSwap struct {
	PID     int
	Pagina  int
	Offset  int
	Tamanio int
}

var TablaSwap []EntradaSwap
var SwapFileMutex sync.Mutex

var TamaniosPorProceso = make(map[int]int)
