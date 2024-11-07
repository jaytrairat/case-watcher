package cfuncs

import "fmt"

// sendAPIRequest sends a POST request to the API with the given message
func sendAPIRequest(message string) error {
	fmt.Printf(message)
	// // Create the HTTP request
	// req, err := http.NewRequest("POST", APIUrl, strings.NewReader("message="+message))
	// if err != nil {
	// 	return fmt.Errorf("failed to create request: %w", err)
	// }

	// // Set the request headers
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Set("x-api-key", APIKey)

	// // Send the request
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return fmt.Errorf("failed to send request: %w", err)
	// }
	// defer resp.Body.Close()

	// // Check the response status
	// if resp.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("received non-OK response: %s", resp.Status)
	// }

	// fmt.Println("API request sent successfully")
	return nil
}
