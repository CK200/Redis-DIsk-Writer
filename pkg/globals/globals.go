package globals

import (
	models "main/pkg/models/configModels"

	"sync"
)

var (
	ApplicationConfig            *models.Config
	ApplicationWaitGroupServices = sync.WaitGroup{}
	Shutdown                     bool
	Workers                      int
	GlobalMutex                  sync.Mutex
	FileCounter                  int
)
