package auth

//go:generate mockgen -destination user_service_mock_test.go -package auth -mock_names Service=MockUserService github.com/prismelabs/prismeanalytics/internal/services/users Service
