package header

func WithMetrics(hs *Service) error {
	return hs.syncer.InitMetrics()
}
