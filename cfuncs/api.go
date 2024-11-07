package cfuncs

import (
	"fmt"
	"net/http"
	"strings"
)

const APIUrl = "http://policeadmin.com:8092/broadcast"
const APIKey = "LDabxoSBFmiedZI2w7o0dVIXbfQnzKV9Bgwy7YNWyfIlB7TWFXPAXS1A1oCN4hNQej7lKxPezvFLYQCtG6f38mAGUw2gKmix71zvw4i5KAJUlHpsPheLF9Q5pgTaUPBi"

func SendAPIRequest(message string) error {
	req, err := http.NewRequest("POST", APIUrl, strings.NewReader("message="+message))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("x-api-key", APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	fmt.Println("API request sent successfully")
	return nil
}
