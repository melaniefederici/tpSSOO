package globals

type IOConfig struct {
	PuertoIO     int    `json:"port_io"`
	IPIO         string `json:"ip_io"`
	IPKernel     string `json:"ip_kernel"`
	PuertoKernel int    `json:"port_kernel"`
	Tipo         string `json:"tipo"`
	Nombre 		 string `json:"nombre"`
}

type PeticionIO struct {
	PID    int    `json:"pid"`
	Tiempo int    `json:"tiempo"`
	Tipo   string `json:"tipo"`
}

// Para el kernel
type RegistroIO struct {
	Nombre string `json:"nombre"`
	IP     string `json:"ip"`
	Puerto int    `json:"puerto"`
	Tipo   string `json:"tipo"`
}

var Config *IOConfig
