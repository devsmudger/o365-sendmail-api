package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const (
	defaultTenantID    = ""
	defaultClientID    = ""
	defaultSenderEmail = ""
	defaultUsername    = ""
	defaultPassword    = ""
	defaultAPIKey      = ""
)

var (
	tenantID    = getEnv("AZURE_TENANT_ID", defaultTenantID)
	clientID    = getEnv("AZURE_CLIENT_ID", defaultClientID)
	senderEmail = getEnv("SENDER_EMAIL", defaultSenderEmail)
	senderUser  = getEnv("SENDER_USERNAME", defaultUsername)
	senderPass  = getEnv("SENDER_PASSWORD", defaultPassword)
	apiKey      = getEnv("API_KEY", defaultAPIKey)
)

// EmailRequest represents the JSON payload for sending email
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// EmailResponse represents the response
type EmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// authMiddleware checks for API key authentication
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if key := r.Header.Get("X-API-Key"); key != apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(EmailResponse{
				Success: false,
				Error:   "Unauthorized: Invalid API key",
			})
			return
		}
		next(w, r)
	}
}

func init() {
	log.Println("SendMail API initialized. Using ROPC auth like sendmail.go.")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getAuthToken(ctx context.Context) (string, error) {
	authority := fmt.Sprintf("https://login.microsoftonline.com/%s", tenantID)
	client, err := public.New(clientID, public.WithAuthority(authority))
	if err != nil {
		return "", err
	}

	scopes := []string{"https://graph.microsoft.com/Mail.Send"}
	result, err := client.AcquireTokenByUsernamePassword(ctx, scopes, senderUser, senderPass)
	if err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func sendEmail(to, subject, body string) error {
	token, err := getAuthToken(context.Background())
	if err != nil {
		return err
	}

	emailData := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]string{
				"contentType": "Text",
				"content":     body,
			},
			"toRecipients": []map[string]interface{}{
				{
					"emailAddress": map[string]string{
						"address": to,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://graph.microsoft.com/v1.0/me/sendMail", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Graph sendMail failed: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// sendEmailHandler handles POST /send-email requests
func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(EmailResponse{
			Success: false,
			Error:   "Only POST requests are allowed",
		})
		return
	}

	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(EmailResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.To == "" || req.Subject == "" || req.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(EmailResponse{
			Success: false,
			Error:   "Missing required fields: to, subject, body",
		})
		return
	}

	err := sendEmail(req.To, req.Subject, req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(EmailResponse{
			Success: false,
			Error:   "Failed to send email: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(EmailResponse{
		Success: true,
		Message: "Email sent successfully to " + req.To,
	})
}

// healthHandler handles GET /health requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/send-email", authMiddleware(sendEmailHandler))

	port := ":8080"
	log.Printf("Starting SendMail API service on %s", port)
	log.Printf("POST /send-email - Send an email (requires X-API-Key header)")
	log.Printf("GET /health - Health check")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
