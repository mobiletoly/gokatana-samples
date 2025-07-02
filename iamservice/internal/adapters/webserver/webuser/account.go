package webuser

import (
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp_echo"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/internal/serverhelp"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/user"
)

// AccountWebHandlers handles user dashboard web requests
type AccountWebHandlers struct {
	authMgm        *usecase.AuthMgm
	userMgm        *usecase.UserMgm
	userProfileMgm *usecase.UserProfileMgm
}

// NewAccountWebHandlers creates a new instance of AccountWebHandlers
func NewAccountWebHandlers(authUC *usecase.AuthMgm, userMgmUC *usecase.UserMgm, userProfileUC *usecase.UserProfileMgm) *AccountWebHandlers {
	return &AccountWebHandlers{
		authMgm:        authUC,
		userMgm:        userMgmUC,
		userProfileMgm: userProfileUC,
	}
}

// AccountLoadHandler renders the user account page
func (h *AccountWebHandlers) AccountLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	userDetails, err := h.userMgm.LoadUserByID(ctx, principal, principal.UserID)
	if err != nil {
		return err
	}
	userProfile, err := h.userProfileMgm.GetUserProfileByUserID(ctx, principal, principal.UserID)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}
	return renderTemplateComponent(c, "Account", user.Account(userDetails, userProfile))
}

// EditAccountLoadHandler renders the edit account form
func (h *AccountWebHandlers) EditAccountLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	userDetails, err := h.userMgm.LoadUserByID(ctx, principal, principal.UserID)
	if err != nil {
		return err
	}
	return renderTemplateComponent(c, "Edit Account", user.EditAccount(userDetails))
}

// UpdateAccountSubmitHandler handles account updates
func (h *AccountWebHandlers) UpdateAccountSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	firstName := strings.TrimSpace(c.FormValue("firstName"))
	lastName := strings.TrimSpace(c.FormValue("lastName"))
	if err = h.userMgm.UpdateUserDetails(ctx, principal, principal.UserID, firstName, lastName); err != nil {
		return err
	}
	return user.AccountUpdateSuccess().Render(ctx, c.Response().Writer)
}

// EditProfileLoadHandler renders the edit profile form
func (h *AccountWebHandlers) EditProfileLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	// Get user profile
	userProfile, err := h.userProfileMgm.GetUserProfileByUserID(ctx, principal, principal.UserID)
	if err != nil {
		return err
	}
	return renderTemplateComponent(c, "Edit Profile", user.EditProfile(userProfile))
}

// UpdateProfileSubmitHandler handles profile updates
func (h *AccountWebHandlers) UpdateProfileSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	// Parse form values
	heightStr := strings.TrimSpace(c.FormValue("height"))
	heightFeetStr := strings.TrimSpace(c.FormValue("height-feet"))
	heightInchesStr := strings.TrimSpace(c.FormValue("height-inches"))
	weightKgStr := strings.TrimSpace(c.FormValue("weight-kg"))
	weightLbsStr := strings.TrimSpace(c.FormValue("weight-lbs"))
	gender := strings.TrimSpace(c.FormValue("gender"))
	birthDateStr := strings.TrimSpace(c.FormValue("birthDate"))
	isMetricStr := strings.TrimSpace(c.FormValue("isMetric"))

	// Create update request
	updateReq := &swagger.UpdateUserProfileRequest{}

	// Determine if user is switching to metric or imperial
	var isMetric bool = true // Default to metric
	if isMetricStr != "" {
		if parsed, err := strconv.ParseBool(isMetricStr); err == nil {
			isMetric = parsed
		}
	}

	if isMetric {
		updateReq.Height, err = parseMetricHeightIntoMillimeters(heightStr)
	} else {
		updateReq.Height, err = parseImperialHeightIntoMillimeters(heightFeetStr, heightInchesStr)
	}
	if err != nil {
		return err
	}

	if isMetric {
		updateReq.Weight, err = parseMetricWeightIntoGrams(weightKgStr)
	} else {
		updateReq.Weight, err = parseImperialWeightIntoGrams(weightLbsStr)
	}

	if birthDateStr != "" {
		if parsedTime, err := time.Parse("2006-01-02", birthDateStr); err == nil {
			updateReq.BirthDate = &openapi_types.Date{Time: parsedTime}
		} else {
			return katapp.NewErr(katapp.ErrInvalidInput, "invalid birth data format")
		}
	}

	if gender != "" {
		s := swagger.UserProfileGender(gender)
		updateReq.Gender = &s
	}
	updateReq.IsMetric = &isMetric

	_, err = h.userProfileMgm.UpdateUserProfileByUserID(ctx, principal, principal.UserID, updateReq)
	if err != nil {
		return err
	}
	return user.ProfileUpdateSuccess().Render(ctx, c.Response().Writer)
}

// ChangePasswordLoadHandler renders the change password form
func (h *AccountWebHandlers) ChangePasswordLoadHandler(c echo.Context) error {
	return renderTemplateComponent(c, "Change Password", user.ChangePassword())
}

// UpdatePasswordSubmitHandler handles password changes
func (h *AccountWebHandlers) UpdatePasswordSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	currentPassword := strings.TrimSpace(c.FormValue("currentPassword"))
	newPassword := strings.TrimSpace(c.FormValue("newPassword"))
	confirmPassword := strings.TrimSpace(c.FormValue("confirmPassword"))
	if newPassword != confirmPassword {
		return katapp.NewErr(katapp.ErrInvalidInput, "New password and confirmation do not match")
	}

	if len(newPassword) < 8 {
		return katapp.NewErr(katapp.ErrInvalidInput, "New password must be at least 8 characters long")
	}

	if err := h.authMgm.ValidateUserPasswordMatches(ctx, principal.UserID, currentPassword); err != nil {
		return err
	}
	err = h.userMgm.ChangeUserPassword(ctx, principal, principal.UserID, newPassword)
	if err != nil {
		return err
	}
	return user.PasswordChangeSuccess().Render(ctx, c.Response().Writer)
}
