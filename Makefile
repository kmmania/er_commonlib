gen_mock_logger:
	mockgen -source=pkg/logger/logger.go \
			-destination=pkg/mocks/logger/mock_logger.go \
			-package=mocks