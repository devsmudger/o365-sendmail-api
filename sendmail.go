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

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func main() {
    tenantID := getEnv("AZURE_TENANT_ID", "")
    clientID := getEnv("AZURE_CLIENT_ID", "")
    username := getEnv("SENDER_USERNAME", "")
    password := getEnv("SENDER_PASSWORD", "")
    recipient := getEnv("RECIPIENT_EMAIL", "")

    if tenantID == "" || clientID == "" || username == "" || password == "" || recipient == "" {
        log.Fatal("Missing required environment variables. Set AZURE_TENANT_ID, AZURE_CLIENT_ID, SENDER_USERNAME, SENDER_PASSWORD, and RECIPIENT_EMAIL.")
    }

    authority := fmt.Sprintf("https://login.microsoftonline.com/%s", tenantID)

    client, err := public.New(clientID, public.WithAuthority(authority))
    if err != nil {
        log.Fatal(err)
    }

    scopes := []string{"https://graph.microsoft.com/Mail.Send"}
    result, err := client.AcquireTokenByUsernamePassword(context.Background(), scopes, username, password)
    if err != nil {
        log.Fatalf("Authentication failed: %v. (Note: This often fails if MFA is enabled)", err)
    }

    emailData := map[string]interface{}{
        "message": map[string]interface{}{
            "subject": "Sent via ROPC Flow",
            "body": map[string]string{
                "contentType": "Text",
                "content":     "Hello! This was sent using a username and password without admin consent.",
            },
            "toRecipients": []map[string]interface{}{
                {
                    "emailAddress": map[string]string{"address": recipient},
                },
            },
        },
    }

    jsonData, err := json.Marshal(emailData)
    if err != nil {
        log.Fatal(err)
    }

    req, err := http.NewRequest("POST", "https://graph.microsoft.com/v1.0/me/sendMail", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Fatal(err)
    }
    req.Header.Set("Authorization", "Bearer "+result.AccessToken)
    req.Header.Set("Content-Type", "application/json")

    httpClient := &http.Client{}
    resp, err := httpClient.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusAccepted {
        fmt.Println("Email sent successfully!")
    } else {
        body, _ := io.ReadAll(resp.Body)
        fmt.Printf("Error: %s\n", string(body))
    }
}
