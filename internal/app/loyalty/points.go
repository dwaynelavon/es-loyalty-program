package loyalty

import "github.com/pkg/errors"

type PointsAction string

const (
	PointsActionReferUser             PointsAction = "ReferUser"
	PointsActionSignUpWithReferral    PointsAction = "SignUpWithReferral"
	PointsActionSignUpWithoutReferral PointsAction = "SignUpWithoutReferral"
)

type PointsMappingService interface {
	Map(action PointsAction) (*uint32, error)
}

type pointsMappingService struct{}

func NewPointsActionMappingService() PointsMappingService {
	return &pointsMappingService{}
}

func (p *pointsMappingService) Map(action PointsAction) (*uint32, error) {
	//TODO: Mapping should live in a store so that values can be changed
	// without re-deploying any applications using the values
	switch action {
	case PointsActionReferUser:
		fallthrough
	case PointsActionSignUpWithReferral:
		return getPointValuePointer(200), nil
	case PointsActionSignUpWithoutReferral:
		return getPointValuePointer(100), nil
	default:
		return nil, errors.Errorf("point action %v not supported", action)
	}
}

func getPointValuePointer(points int) *uint32 {
	p := uint32(points)
	return &p
}
