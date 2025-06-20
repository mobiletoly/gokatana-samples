// Code generated by gobetter; DO NOT EDIT.

package repo

import (
	"time"
)

func NewAuthUserEntityBuilder() AuthUserEntity_Builder_ID {
	return AuthUserEntity_Builder_ID{root: &AuthUserEntity{}}
}

type AuthUserEntity_Builder_ID struct {
	root *AuthUserEntity
}

type AuthUserEntity_Builder_Email struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_ID) ID(arg *string) AuthUserEntity_Builder_Email {
	b.root.ID = arg
	return AuthUserEntity_Builder_Email{root: b.root}
}

type AuthUserEntity_Builder_PasswordHash struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_Email) Email(arg string) AuthUserEntity_Builder_PasswordHash {
	b.root.Email = arg
	return AuthUserEntity_Builder_PasswordHash{root: b.root}
}

type AuthUserEntity_Builder_FirstName struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_PasswordHash) PasswordHash(arg string) AuthUserEntity_Builder_FirstName {
	b.root.PasswordHash = arg
	return AuthUserEntity_Builder_FirstName{root: b.root}
}

type AuthUserEntity_Builder_LastName struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_FirstName) FirstName(arg string) AuthUserEntity_Builder_LastName {
	b.root.FirstName = arg
	return AuthUserEntity_Builder_LastName{root: b.root}
}

type AuthUserEntity_Builder_IsActive struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_LastName) LastName(arg string) AuthUserEntity_Builder_IsActive {
	b.root.LastName = arg
	return AuthUserEntity_Builder_IsActive{root: b.root}
}

type AuthUserEntity_Builder_EmailVerified struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_IsActive) IsActive(arg bool) AuthUserEntity_Builder_EmailVerified {
	b.root.IsActive = arg
	return AuthUserEntity_Builder_EmailVerified{root: b.root}
}

type AuthUserEntity_Builder_CreatedAt struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_EmailVerified) EmailVerified(arg bool) AuthUserEntity_Builder_CreatedAt {
	b.root.EmailVerified = arg
	return AuthUserEntity_Builder_CreatedAt{root: b.root}
}

type AuthUserEntity_Builder_UpdatedAt struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_CreatedAt) CreatedAt(arg *time.Time) AuthUserEntity_Builder_UpdatedAt {
	b.root.CreatedAt = arg
	return AuthUserEntity_Builder_UpdatedAt{root: b.root}
}

type AuthUserEntity_Builder_GobFinalizer struct {
	root *AuthUserEntity
}

func (b AuthUserEntity_Builder_UpdatedAt) UpdatedAt(arg *time.Time) AuthUserEntity_Builder_GobFinalizer {
	b.root.UpdatedAt = arg
	return AuthUserEntity_Builder_GobFinalizer{root: b.root}
}

func (b AuthUserEntity_Builder_GobFinalizer) Build() *AuthUserEntity {
	return b.root
}

func NewAuthRoleEntityBuilder() AuthRoleEntity_Builder_ID {
	return AuthRoleEntity_Builder_ID{root: &AuthRoleEntity{}}
}

type AuthRoleEntity_Builder_ID struct {
	root *AuthRoleEntity
}

type AuthRoleEntity_Builder_Name struct {
	root *AuthRoleEntity
}

func (b AuthRoleEntity_Builder_ID) ID(arg *int) AuthRoleEntity_Builder_Name {
	b.root.ID = arg
	return AuthRoleEntity_Builder_Name{root: b.root}
}

type AuthRoleEntity_Builder_Description struct {
	root *AuthRoleEntity
}

func (b AuthRoleEntity_Builder_Name) Name(arg string) AuthRoleEntity_Builder_Description {
	b.root.Name = arg
	return AuthRoleEntity_Builder_Description{root: b.root}
}

type AuthRoleEntity_Builder_CreatedAt struct {
	root *AuthRoleEntity
}

func (b AuthRoleEntity_Builder_Description) Description(arg *string) AuthRoleEntity_Builder_CreatedAt {
	b.root.Description = arg
	return AuthRoleEntity_Builder_CreatedAt{root: b.root}
}

type AuthRoleEntity_Builder_UpdatedAt struct {
	root *AuthRoleEntity
}

func (b AuthRoleEntity_Builder_CreatedAt) CreatedAt(arg *time.Time) AuthRoleEntity_Builder_UpdatedAt {
	b.root.CreatedAt = arg
	return AuthRoleEntity_Builder_UpdatedAt{root: b.root}
}

type AuthRoleEntity_Builder_GobFinalizer struct {
	root *AuthRoleEntity
}

func (b AuthRoleEntity_Builder_UpdatedAt) UpdatedAt(arg *time.Time) AuthRoleEntity_Builder_GobFinalizer {
	b.root.UpdatedAt = arg
	return AuthRoleEntity_Builder_GobFinalizer{root: b.root}
}

func (b AuthRoleEntity_Builder_GobFinalizer) Build() *AuthRoleEntity {
	return b.root
}

func NewAuthUserRoleEntityBuilder() AuthUserRoleEntity_Builder_UserID {
	return AuthUserRoleEntity_Builder_UserID{root: &AuthUserRoleEntity{}}
}

type AuthUserRoleEntity_Builder_UserID struct {
	root *AuthUserRoleEntity
}

type AuthUserRoleEntity_Builder_RoleID struct {
	root *AuthUserRoleEntity
}

func (b AuthUserRoleEntity_Builder_UserID) UserID(arg string) AuthUserRoleEntity_Builder_RoleID {
	b.root.UserID = arg
	return AuthUserRoleEntity_Builder_RoleID{root: b.root}
}

type AuthUserRoleEntity_Builder_AssignedAt struct {
	root *AuthUserRoleEntity
}

func (b AuthUserRoleEntity_Builder_RoleID) RoleID(arg int) AuthUserRoleEntity_Builder_AssignedAt {
	b.root.RoleID = arg
	return AuthUserRoleEntity_Builder_AssignedAt{root: b.root}
}

type AuthUserRoleEntity_Builder_AssignedBy struct {
	root *AuthUserRoleEntity
}

func (b AuthUserRoleEntity_Builder_AssignedAt) AssignedAt(arg *time.Time) AuthUserRoleEntity_Builder_AssignedBy {
	b.root.AssignedAt = arg
	return AuthUserRoleEntity_Builder_AssignedBy{root: b.root}
}

type AuthUserRoleEntity_Builder_GobFinalizer struct {
	root *AuthUserRoleEntity
}

func (b AuthUserRoleEntity_Builder_AssignedBy) AssignedBy(arg *string) AuthUserRoleEntity_Builder_GobFinalizer {
	b.root.AssignedBy = arg
	return AuthUserRoleEntity_Builder_GobFinalizer{root: b.root}
}

func (b AuthUserRoleEntity_Builder_GobFinalizer) Build() *AuthUserRoleEntity {
	return b.root
}
