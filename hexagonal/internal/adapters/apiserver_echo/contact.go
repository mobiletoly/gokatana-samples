package apiserver_echo

import (
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/usecase"
	"github.com/mobiletoly/gokatana/kathttp_echo"
	"net/http"

	"github.com/labstack/echo/v4"
)

func getContactByIDRoute(uc *usecase.Contact) func(c echo.Context) error {
	return func(c echo.Context) error {
		ID := c.Param("id")
		ctx := c.Request().Context()
		contact, err := uc.LoadContactByID(ctx, ID)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusOK, contact)
	}
}

func getAllContactsRoute(uc *usecase.Contact) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		contacts, err := uc.LoadAllContacts(ctx)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusOK, contacts)
	}
}

func addContactRoute(uc *usecase.Contact) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var addContact model.AddContact
		if err := c.Bind(&addContact); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		if contact, err := uc.AddContact(ctx, &addContact); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusCreated, contact)
		}
	}
}
