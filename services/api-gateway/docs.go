// Package main provides the API Gateway for Chthon ShortLink service
//
// This is the main entry point for the Chthon ShortLink microservices platform.
// The API Gateway serves as a single point of entry for all client requests,
// handling authentication, routing, and load balancing.
//
//	@title			Chthon ShortLink API Gateway
//	@version		1.0.0
//	@description	Enterprise-grade URL shortening microservices platform
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.url	http://localhost:8080/support
//	@contact.email	support@shortlink.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	JWT
//	@in							header
//	@name						Authorization
//	@description				JWT token for authentication. Format: Bearer {token}
//
//	@schemes	http https
//	@produce	json
//	@consumes	json
//
//	@tag.name			Auth
//	@tag.description	Authentication and authorization endpoints
//
//	@tag.name			ShortLinks
//	@tag.description	URL shortening operations
//
//	@tag.name			Analytics
//	@tag.description	Analytics and statistics endpoints
//
//	@tag.name			Users
//	@tag.description	User management operations
//
//	@tag.name			Health
//	@tag.description	System health and monitoring
package main
