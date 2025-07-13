package intgr_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// MockEmail represents an email that was sent during testing
type MockEmail struct {
	To          string    `json:"to"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	ContentType string    `json:"contentType"`
	SentAt      time.Time `json:"sentAt"`
}

// clearMockEmails removes all mock email files to start fresh
func clearMockEmails() {
	// Remove the emails file
	_ = os.Remove("test-emails/emails.json")
}

// getAllMockEmails reads all mock emails from the file
func getAllMockEmails() ([]MockEmail, error) {
	file, err := os.Open("test-emails/emails.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []MockEmail{}, nil // Return empty slice if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	var emails []MockEmail
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var email MockEmail
		if err := json.Unmarshal(scanner.Bytes(), &email); err != nil {
			return nil, fmt.Errorf("failed to parse email JSON: %w", err)
		}
		emails = append(emails, email)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read emails file: %w", err)
	}

	return emails, nil
}

// getLastMockEmail returns the most recently sent mock email
func getLastMockEmail() (*MockEmail, error) {
	emails, err := getAllMockEmails()
	if err != nil {
		return nil, err
	}
	if len(emails) == 0 {
		return nil, fmt.Errorf("no mock emails found")
	}
	return &emails[len(emails)-1], nil
}

// getMockEmailsTo returns all mock emails sent to a specific address
func getMockEmailsTo(to string) ([]MockEmail, error) {
	emails, err := getAllMockEmails()
	if err != nil {
		return nil, err
	}

	var filtered []MockEmail
	for _, email := range emails {
		if email.To == to {
			filtered = append(filtered, email)
		}
	}
	return filtered, nil
}

// getMockEmailCount returns the total number of mock emails sent
func getMockEmailCount() (int, error) {
	emails, err := getAllMockEmails()
	if err != nil {
		return 0, err
	}
	return len(emails), nil
}

// extractSixDigitCode extracts a 6-digit confirmation code from email body
func extractSixDigitCode(emailBody string) string {
	// Look for 6-digit code pattern in the email body
	re := regexp.MustCompile(`\b\d{6}\b`)
	matches := re.FindAllString(emailBody, -1)

	// Return the first 6-digit code found
	for _, match := range matches {
		if len(match) == 6 {
			return match
		}
	}
	return ""
}

// extractConfirmationURL extracts the confirmation URL from email body
func extractConfirmationURL(emailBody string) string {
	// Look for the confirmation URL pattern in the email body
	// Handle both API and web URLs, and both regular & and HTML-encoded &amp; in URLs
	re := regexp.MustCompile(`/(?:api/v1/auth|web/user/auth)/confirm-email\?userId=[^"&\s]+(?:&amp;|&)code=[^"&\s]+`)
	matches := re.FindString(emailBody)
	// Replace HTML-encoded ampersands with regular ones
	matches = strings.ReplaceAll(matches, "&amp;", "&")
	return strings.TrimSpace(matches)
}

// extractUserIDFromConfirmationURL extracts the user ID from a confirmation URL
func extractUserIDFromConfirmationURL(confirmationURL string) string {
	re := regexp.MustCompile(`userId=([^&]+)`)
	matches := re.FindStringSubmatch(confirmationURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractCodeFromConfirmationURL extracts the confirmation code from a confirmation URL
func extractCodeFromConfirmationURL(confirmationURL string) string {
	re := regexp.MustCompile(`code=([^&]+)`)
	matches := re.FindStringSubmatch(confirmationURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractConfirmationCode extracts confirmation code from email body (handles both 6-digit codes and URL tokens)
func extractConfirmationCode(emailBody string) string {
	// First try to extract 6-digit code (mobile)
	sixDigitCode := extractSixDigitCode(emailBody)
	if len(sixDigitCode) == 6 {
		return sixDigitCode
	}

	// If no 6-digit code, try to extract from URL (web)
	confirmationURL := extractConfirmationURL(emailBody)
	if confirmationURL != "" {
		return extractCodeFromConfirmationURL(confirmationURL)
	}

	return ""
}

// validateEmailContent performs basic validation on email content
func validateEmailContent(email *MockEmail, expectedRecipient string, expectedSubjectContains string) bool {
	if email == nil {
		return false
	}

	if email.To != expectedRecipient {
		return false
	}

	if !strings.Contains(email.Subject, expectedSubjectContains) {
		return false
	}

	if email.ContentType != "text/html" {
		return false
	}

	if !strings.Contains(email.Body, "IAMService") {
		return false
	}

	return true
}

// validateWebEmailContent validates web-specific email content
func validateWebEmailContent(email *MockEmail, expectedRecipient string, expectedUserName string) bool {
	if !validateEmailContent(email, expectedRecipient, "Confirm Your Email Address") {
		return false
	}

	// Check for web-specific content
	if !strings.Contains(email.Body, "Confirm Email Address") { // Button text
		return false
	}

	if !strings.Contains(email.Body, expectedUserName) {
		return false
	}

	// Should contain a confirmation URL
	confirmationURL := extractConfirmationURL(email.Body)
	return confirmationURL != ""
}

// validateMobileEmailContent validates mobile-specific email content
func validateMobileEmailContent(email *MockEmail, expectedRecipient string, expectedUserName string, expectedPlatform string) bool {
	if !validateEmailContent(email, expectedRecipient, "Your Confirmation Code") {
		return false
	}

	// Check for mobile-specific content (case-insensitive platform check)
	if !strings.Contains(strings.ToLower(email.Subject), strings.ToLower(expectedPlatform)) {
		return false
	}

	if !strings.Contains(email.Body, expectedUserName) {
		return false
	}

	if !strings.Contains(email.Body, strings.ToLower(expectedPlatform)) {
		return false
	}

	// Should contain a 6-digit code
	code := extractSixDigitCode(email.Body)
	return len(code) == 6
}

// waitForMockEmail waits for a mock email to be sent (useful for async operations)
func waitForMockEmail(expectedCount int, timeoutSeconds int) error {
	for i := 0; i < timeoutSeconds*10; i++ { // Check every 100ms
		count, err := getMockEmailCount()
		if err != nil {
			return err
		}
		if count >= expectedCount {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for mock email (expected %d emails)", expectedCount)
}
