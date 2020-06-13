package dependency

import "github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"

func NewPointsMappingService() loyalty.PointsMappingService {
	return loyalty.NewPointsActionMappingService()
}
