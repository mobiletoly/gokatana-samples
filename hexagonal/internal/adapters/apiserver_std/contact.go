package apiserver_std

import (
	"encoding/json"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/usecase"
	"github.com/mobiletoly/gokatana/kathttp_std"
	"net/http"
)

func getContactByIDRoute(uc *usecase.Contact) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read param "id"
		ID := r.PathValue("id")
		ctx := r.Context()
		contact, err := uc.LoadContactByID(ctx, ID)
		if err != nil {
			kathttp_std.ReportHTTPError(w, err)
			return
		}
		if err := json.NewEncoder(w).Encode(contact); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func getAllContactsRoute(uc *usecase.Contact) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contacts, err := uc.LoadAllContacts(ctx)
		if err != nil {
			kathttp_std.ReportHTTPError(w, err)
			return
		}
		if err := json.NewEncoder(w).Encode(contacts); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func addContactRoute(uc *usecase.Contact) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var addContact model.AddContact
		if err := kathttp_std.Bind(r, &addContact); err != nil {
			kathttp_std.ReportHTTPError(w, err)
			return
		}
		if contact, err := uc.AddContact(ctx, &addContact); err != nil {
			kathttp_std.ReportHTTPError(w, err)
		} else {
			if err := json.NewEncoder(w).Encode(contact); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}
	}
}
