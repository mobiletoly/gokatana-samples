// Code generated by gobetter; DO NOT EDIT.

package swagger

import (
	"github.com/go-openapi/strfmt"
)

func NewUserProfileBuilder() UserProfile_Builder_ID {
	return UserProfile_Builder_ID{root: &UserProfile{}}
}

type UserProfile_Builder_ID struct {
	root *UserProfile
}

type UserProfile_Builder_Email struct {
	root *UserProfile
}

func (b UserProfile_Builder_ID) ID(arg *string) UserProfile_Builder_Email {
	b.root.ID = arg
	return UserProfile_Builder_Email{root: b.root}
}

type UserProfile_Builder_FirstName struct {
	root *UserProfile
}

func (b UserProfile_Builder_Email) Email(arg *strfmt.Email) UserProfile_Builder_FirstName {
	b.root.Email = arg
	return UserProfile_Builder_FirstName{root: b.root}
}

type UserProfile_Builder_LastName struct {
	root *UserProfile
}

func (b UserProfile_Builder_FirstName) FirstName(arg *string) UserProfile_Builder_LastName {
	b.root.FirstName = arg
	return UserProfile_Builder_LastName{root: b.root}
}

type UserProfile_Builder_CreatedAt struct {
	root *UserProfile
}

func (b UserProfile_Builder_LastName) LastName(arg *string) UserProfile_Builder_CreatedAt {
	b.root.LastName = arg
	return UserProfile_Builder_CreatedAt{root: b.root}
}

type UserProfile_Builder_UpdatedAt struct {
	root *UserProfile
}

func (b UserProfile_Builder_CreatedAt) CreatedAt(arg *strfmt.DateTime) UserProfile_Builder_UpdatedAt {
	b.root.CreatedAt = arg
	return UserProfile_Builder_UpdatedAt{root: b.root}
}

type UserProfile_Builder_GobFinalizer struct {
	root *UserProfile
}

func (b UserProfile_Builder_UpdatedAt) UpdatedAt(arg *strfmt.DateTime) UserProfile_Builder_GobFinalizer {
	b.root.UpdatedAt = arg
	return UserProfile_Builder_GobFinalizer{root: b.root}
}

func (b UserProfile_Builder_GobFinalizer) Build() *UserProfile {
	return b.root
}
