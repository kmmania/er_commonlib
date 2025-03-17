gen_mock_logger:
	mockgen -source=pkg/logger/logger.go \
			-destination=pkg/mocks/logger/mock_logger.go \
			-package=mocks

gen_mock_db:
	mockgen -source=pkg/db/db.go \
			-destination=pkg/mocks/db/mock_db.go \
			-package=mocks