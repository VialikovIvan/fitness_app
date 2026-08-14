package service

type Service struct {

}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CheckHealth() bool {
	return true
}