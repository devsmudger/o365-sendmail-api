# O365 SendMail API and Script

This repository provides two Go-based mail sender tools for Microsoft 365 / Office 365:

1. `sendmail.go` - a command-line script using ROPC auth to send a single email.
2. `sendmail_api.go` - an HTTP API service that accepts JSON and sends mail via Microsoft Graph.

## Files

- `sendmail.go` - simple send-mail script
- `sendmail_api.go` - HTTP API server
- `go.mod` / `go.sum` - Go module dependencies
- `.env.example` - environment variable template
- `.gitignore` - ignore local secrets and build outputs

## Environment Variables

Set these variables before running either tool:

\`\`\`bash
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_CLIENT_ID="your-client-id"
export SENDER_EMAIL="your-sender-email@example.com"
export SENDER_USERNAME="your-username@example.com"
export SENDER_PASSWORD="your-password"
export API_KEY="your-secure-api-key"
\`\`\`

For the script, you also need:

\`\`\`bash
export RECIPIENT_EMAIL="recipient@example.com"
\`\`\`

## Run the script

From `/opt/go/O365SendMail`:

\`\`\`bash
go run sendmail.go
\`\`\`

The script reads credentials from environment variables and sends a single email to `RECIPIENT_EMAIL`.

## Build and run the API

From `/opt/go/O365SendMail`:

\`\`\`bash
go build -o sendmail_api sendmail_api.go
export $(cat .env | xargs)
./sendmail_api
\`\`\`

The API listens on `http://localhost:8080`.

## API Endpoints

### Health check

\`\`\`bash
curl http://localhost:8080/health
\`\`\`

### Send email

\`\`\`bash
curl -X POST http://localhost:8080/send-email \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{"to":"recipient@example.com","subject":"Hello","body":"Test email"}'
\`\`\`

Successful response:

\`\`\`json
{
  "success": true,
  "message": "Email sent successfully to recipient@example.com"
}
\`\`\`

## Security notes

- Do not commit `.env` or any credentials to git.
- Use a strong `API_KEY` and keep it secret.
- In production, use HTTPS and do not expose the API without additional protections.

## Notes

- `sendmail.go` uses direct username/password auth for a single email.
- `sendmail_api.go` provides a reusable endpoint for application integration.
