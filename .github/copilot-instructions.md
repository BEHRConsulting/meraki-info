<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# Meraki Routes Backup - Copilot Instructions

This is a Golang application that backs up Meraki network routes using the Meraki Dashboard API.

## Project Structure
- `main.go`: Entry point of the application
- `internal/config/`: Configuration management (CLI flags and environment variables)
- `internal/logger/`: Structured logging with slog
- `internal/meraki/`: Meraki API client with OAuth2 support
- `internal/output/`: Output formatters (text, JSON, XML, CSV)

## Code Guidelines
- Use structured logging with `log/slog`
- Follow Go best practices and idioms
- Include comprehensive error handling
- Write unit tests for all packages
- Use dependency injection for better testability
- Follow the repository pattern for API interactions

## API Authentication
- Primary method: API Key authentication
- OAuth2 support available for production environments
- Rate limiting consideration for API calls

## Error Handling
- Wrap errors with meaningful context
- Use structured logging for error reporting
- Graceful degradation when possible
- Clear error messages for end users

## Testing
- Unit tests for all packages
- Mock HTTP responses for API testing
- Table-driven tests where appropriate
- Test coverage for error scenarios
