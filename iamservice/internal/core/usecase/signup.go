package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase/internal"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/email"
	"github.com/mobiletoly/gokatana/katapp"
	"strings"
	"time"
)

// SignUp creates a new user account
func (a *AuthMgm) SignUp(ctx context.Context, req *swagger.SignupRequest) (*swagger.SignupResponse, error) {
	katapp.Logger(ctx).Info("signing up user", "email", string(req.Email), "tenantID", req.TenantId)
	if err := a.validateSignupRequest(req); err != nil {
		return nil, err
	}

	user, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) (*model.AuthUser, error) {
		tenantID := req.TenantId
		if err := internal.EnsureTenantExistsById(ctx, a.authUserPersist, tx, tenantID); err != nil {
			return nil, err
		}

		existingUser, err := a.authUserPersist.GetUserByEmail(ctx, tx, string(req.Email), tenantID)
		if err != nil {
			katapp.Logger(ctx).Error("failed to check existing user", "email", string(req.Email), "tenantID", tenantID, "error", err)
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to check existing user")
		}
		if existingUser != nil {
			if existingUser.EmailVerified {
				// User exists and email is verified - cannot sign up again
				katapp.Logger(ctx).Warn("user already exists with verified email", "email", string(req.Email), "tenantID", tenantID)
				return nil, katapp.NewErr(katapp.ErrDuplicate, "user with this email already exists")
			} else {
				// User exists but email is not verified - allow re-signup by updating existing user
				katapp.Logger(ctx).Info("user exists with unverified email, allowing re-signup", "email", string(req.Email), "tenantID", tenantID, "userID", existingUser.ID)
				// We'll update the existing user instead of creating a new one
			}
		}

		hashedPassword, err := internal.HashPassword(req.Password)
		if err != nil {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to hash password")
		}

		var user *model.AuthUser
		if existingUser != nil && !existingUser.EmailVerified {
			// Delete existing unverified user and create a new one
			katapp.Logger(ctx).Info("deleting existing unverified user for re-signup", "userID", existingUser.ID, "email", string(req.Email))
			err = a.authUserPersist.DeleteUser(ctx, tx, existingUser.ID)
			if err != nil {
				katapp.Logger(ctx).Error("failed to delete existing unverified user", "userID", existingUser.ID, "error", err)
				return nil, katapp.NewErr(katapp.ErrInternal, "failed to delete existing unverified user")
			}
		}

		// Create new user (either first time or replacing unverified user)
		signupReq := *req
		signupReq.Password = hashedPassword
		user, err = a.authUserPersist.CreateUser(ctx, tx, &signupReq, tenantID)
		if err != nil {
			return nil, err
		}

		// Assign default 'user' role to new user
		err = a.authUserPersist.AssignUserRole(ctx, tx, user.ID, "user")
		if err != nil {
			katapp.Logger(ctx).Warn("failed to assign default role to user", "userID", user.ID, "error", err)
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to assign default role")
		}

		if existingUser != nil && !existingUser.EmailVerified {
			katapp.Logger(ctx).Info("replaced existing unverified user", "oldUserID", existingUser.ID, "newUserID", user.ID, "email", string(req.Email))
		} else {
			katapp.Logger(ctx).Info("created new user", "userID", user.ID, "email", string(req.Email))
		}

		// Generate confirmation token/code and hash it
		var tokenForEmail string
		var tokenHash string

		if req.Source == "web" {
			// For web: generate long token
			tokenForEmail, err = a.generateEmailConfirmationToken()
			if err != nil {
				katapp.Logger(ctx).Error("failed to generate email confirmation token", "userID", user.ID, "error", err)
				return nil, katapp.NewErr(katapp.ErrInternal, "failed to generate email confirmation token")
			}
			tokenHash = a.hashToken(user.ID, tokenForEmail)
		} else {
			// For mobile: generate 6-digit code
			tokenForEmail = a.generateSixDigitCode()
			tokenHash = a.hashToken(user.ID, tokenForEmail)
		}

		// Create email confirmation token in database (expires in 24 hours)
		expiresAt := time.Now().Add(24 * time.Hour)
		_, err = a.authUserPersist.CreateEmailConfirmationToken(
			ctx, tx, user.ID, user.Email, tokenHash,
			string(req.Source),
			expiresAt)
		if err != nil {
			katapp.Logger(ctx).Error("failed to create email confirmation token",
				"userID", user.ID,
				"source", req.Source,
				"error", err)
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to create email confirmation token")
		}

		// Send confirmation email based on source
		err = a.sendConfirmationEmail(ctx, user, tokenForEmail, string(req.Source))
		if err != nil {
			katapp.Logger(ctx).Error("failed to send confirmation email",
				"userID", user.ID,
				"source", req.Source,
				"error", err)
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to send confirmation email")
		}

		katapp.Logger(ctx).Info("email confirmation token created and email sent",
			"userID", user.ID,
			"source", req.Source)
		return user, nil
	})

	if err != nil {
		return nil, err
	}

	// Return a response indicating that email confirmation is required
	message := "User account created successfully. Please check your email to confirm your account."
	return &swagger.SignupResponse{
		Message: message,
		Email:   req.Email,
		UserId:  user.ID,
	}, nil
}

func (a *AuthMgm) validateSignupRequest(req *swagger.SignupRequest) error {
	if req.Email == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "email is required")
	}
	if req.Password == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "password is required")
	}
	if len(req.Password) < 8 {
		return katapp.NewErr(katapp.ErrInvalidInput, "password must be at least 8 characters")
	}
	if req.FirstName == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "first name is required")
	}
	if req.LastName == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "last name is required")
	}
	if req.TenantId == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "tenant ID is required")
	}
	if req.Source == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "source is required")
	}

	// Basic email validation
	if !strings.Contains(string(req.Email), "@") {
		return katapp.NewErr(katapp.ErrInvalidInput, "invalid email format")
	}

	return nil
}

// ConfirmEmail confirms a user's email address using the confirmation token/code
func (a *AuthMgm) ConfirmEmail(ctx context.Context, userID string, code string) error {
	katapp.Logger(ctx).Info("confirming email", "userID", userID)

	if userID == "" || code == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "user ID and confirmation code are required")
	}

	err := a.txPort.Run(ctx, func(tx pgx.Tx) error {
		// Hash the provided code with user ID for database lookup
		codeHash := a.hashToken(userID, code)

		// Get the confirmation token by user ID and hash
		confirmationToken, err := a.authUserPersist.GetEmailConfirmationTokenByUserIDAndHash(ctx, tx, userID, codeHash)
		if err != nil {
			katapp.Logger(ctx).Error("failed to get email confirmation token", "userID", userID, "error", err)
			return katapp.NewErr(katapp.ErrInternal, "failed to get confirmation token")
		}

		if confirmationToken == nil {
			return katapp.NewErr(katapp.ErrNotFound, "invalid confirmation code")
		}

		// Check if token is valid
		if !confirmationToken.IsValid() {
			if confirmationToken.IsExpired() {
				return katapp.NewErr(katapp.ErrInvalidInput, "confirmation token has expired")
			}
			if confirmationToken.IsUsed() {
				return katapp.NewErr(katapp.ErrInvalidInput, "confirmation token has already been used")
			}
		}

		// Mark token as used
		err = a.authUserPersist.MarkEmailConfirmationTokenAsUsed(ctx, tx, confirmationToken.ID)
		if err != nil {
			katapp.Logger(ctx).Error("failed to mark confirmation token as used", "tokenID", confirmationToken.ID, "error", err)
			return katapp.NewErr(katapp.ErrInternal, "failed to mark token as used")
		}

		// Set user email as verified
		err = a.authUserPersist.SetUserEmailVerified(ctx, tx, confirmationToken.UserID, true)
		if err != nil {
			katapp.Logger(ctx).Error("failed to set user email as verified", "userID", confirmationToken.UserID, "error", err)
			return katapp.NewErr(katapp.ErrInternal, "failed to verify email")
		}

		katapp.Logger(ctx).Info("email confirmed successfully", "userID", confirmationToken.UserID)
		return nil
	})

	return err
}

// Email confirmation helper methods

// sendConfirmationEmail sends platform-specific confirmation emails
func (a *AuthMgm) sendConfirmationEmail(ctx context.Context, user *model.AuthUser, token string, source string) error {
	switch source {
	case "web":
		return a.sendWebConfirmationEmail(ctx, user, token)
	case "android", "ios":
		return a.sendMobileConfirmationEmail(ctx, user, token, source)
	default:
		return katapp.NewErr(katapp.ErrInvalidInput, "invalid source platform")
	}
}

// sendWebConfirmationEmail sends a clickable confirmation link for web users
func (a *AuthMgm) sendWebConfirmationEmail(ctx context.Context, user *model.AuthUser, token string) error {
	baseURL := a.serverConfig.Domain
	confirmationURL := fmt.Sprintf("%s/api/v1/auth/confirm-email?userId=%s&code=%s", baseURL, user.ID, token)

	data := &email.WebConfirmationData{
		User:            user,
		ConfirmationURL: confirmationURL,
		ExpiresIn:       "24 hours",
	}

	// Render the email template
	var buf strings.Builder
	err := email.WebConfirmation(data).Render(ctx, &buf)
	if err != nil {
		return katapp.NewErr(katapp.ErrInternal, "failed to render email template")
	}

	// Create mail content
	mailContent := outport.NewMailContentBuilder().
		ContentType("text/html").
		Title("Confirm Your Email Address - IAMService").
		Body(buf.String()).
		Build()

	// Send email
	err = a.mailer.SendEmail(ctx, user.Email, mailContent)
	if err != nil {
		return katapp.NewErr(katapp.ErrInternal, "failed to send confirmation email")
	}

	katapp.Logger(ctx).Info("web confirmation email sent", "userID", user.ID, "email", user.Email)
	return nil
}

// sendMobileConfirmationEmail sends a 6-digit confirmation code for mobile users
func (a *AuthMgm) sendMobileConfirmationEmail(ctx context.Context, user *model.AuthUser, confirmationCode string, platform string) error {
	// Use the confirmation code that was already generated and stored in the database

	data := &email.MobileConfirmationData{
		User:             user,
		ConfirmationCode: confirmationCode,
		ExpiresIn:        "24 hours",
		Platform:         platform,
	}

	// Render the email template
	var buf strings.Builder
	err := email.MobileConfirmation(data).Render(ctx, &buf)
	if err != nil {
		return katapp.NewErr(katapp.ErrInternal, "failed to render email template")
	}

	// Create mail content
	mailContent := outport.NewMailContentBuilder().
		ContentType("text/html").
		Title(fmt.Sprintf("Your Confirmation Code - IAMService (%s)", strings.Title(platform))).
		Body(buf.String()).
		Build()

	// Send email
	err = a.mailer.SendEmail(ctx, user.Email, mailContent)
	if err != nil {
		return katapp.NewErr(katapp.ErrInternal, "failed to send confirmation email")
	}

	katapp.Logger(ctx).Info("mobile confirmation email sent", "userID", user.ID, "email", user.Email, "platform", platform, "code", confirmationCode)
	return nil
}

// generateSixDigitCode generates a random 6-digit confirmation code
func (a *AuthMgm) generateSixDigitCode() string {
	b := make([]byte, 3) // 3 bytes = 24 bits, enough for 6 digits
	_, _ = rand.Read(b)
	// Convert bytes to a number and ensure it's 6 digits
	num := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", num%1000000)
}

// generateEmailConfirmationToken generates a secure random token for email confirmation
func (a *AuthMgm) generateEmailConfirmationToken() (string, error) {
	b := make([]byte, 32) // 256-bit token
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b), nil
}
