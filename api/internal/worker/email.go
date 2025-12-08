package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"p2m-lite/config"
)

type Sender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type To struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type EmailPayload struct {
	Sender      Sender `json:"sender"`
	To          []To   `json:"to"`
	Subject     string `json:"subject"`
	HtmlContent string `json:"htmlContent"`
}

func sendEmail(cfg *config.Config, recorder string, ph, turbidity, lat, lon float64) error {
	mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%f,%f", lat, lon)
	url := "https://api.brevo.com/v3/smtp/email"
	adminEmail := "rishabh.kumar.pro@gmail.com"

	payload := EmailPayload{
		Sender: Sender{
			Name:  "P2M Bot",
			Email: "p2m@040203.xyz",
		},
		To: []To{
			{Email: adminEmail, Name: "Admin"},
		},
		Subject: fmt.Sprintf("Alert: Low Water Quality for Recorder %s", recorder),
		HtmlContent: fmt.Sprintf(`
			<html>
				<body>
					<h1>Low Water Quality Detected</h1>
					<p><strong>Recorder:</strong> %s</p>
					<p><strong>Average pH:</strong> %.2f</p>
					<p><strong>Average Turbidity:</strong> %.2f</p>
					<p><strong>Location:</strong> <a href="%s">View on Google Maps</a> (Lat: %f, Lon: %f)</p>
					<p>Please investigate immediately.</p>
				</body>
			</html>
		`, recorder, ph, turbidity, mapsLink, lat, lon),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", cfg.BrevoAPIKey)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to send email, status code: %d", resp.StatusCode)
	}

	return nil
}
