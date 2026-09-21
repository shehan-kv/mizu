package contract

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

func newTestProject(t *testing.T) *project.Project {
	t.Helper()

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatalf("failed to create project ID: %v", err)
	}

	name, err := project.NewName("Test Project")
	if err != nil {
		t.Fatalf("failed to create project name: %v", err)
	}

	status, err := project.NewStatus("started")
	if err != nil {
		t.Fatalf("failed to create project status: %v", err)
	}

	createdBy := iam.UserID("admin-1")

	prj, err := project.NewProject(
		projectID,
		name,
		status,
		createdBy,
		[]iam.UserID{
			iam.UserID("client-1"),
			iam.UserID("staff-1"),
			iam.UserID("admin-1"),
		},
		time.Now(),
	)
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	return prj
}

func newTestUser(
	t *testing.T,
	id iam.UserID,
	role iam.Role,
) *iam.User {
	t.Helper()

	name, err := iam.NewName("Test", "User")
	if err != nil {
		t.Fatalf("failed to create user name: %v", err)
	}

	email, err := iam.NewEmail(string(id) + "@example.com")
	if err != nil {
		t.Fatalf("failed to create user email: %v", err)
	}

	return iam.NewSystemUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		time.Now(),
	)
}

func TestServiceValidateSignatories(t *testing.T) {
	service := NewService()

	t.Run("accepts client and team member who are project members", func(t *testing.T) {
		prj := newTestProject(t)

		client := newTestUser(
			t,
			iam.UserID("client-1"),
			iam.RoleClient,
		)

		staff := newTestUser(
			t,
			iam.UserID("staff-1"),
			iam.RoleStaff,
		)

		err := service.ValidateSignatories(
			[]*iam.User{client, staff},
			prj,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("accepts client and administrator who are project members", func(t *testing.T) {
		prj := newTestProject(t)

		client := newTestUser(
			t,
			iam.UserID("client-1"),
			iam.RoleClient,
		)

		admin := newTestUser(
			t,
			iam.UserID("admin-1"),
			iam.RoleAdministrator,
		)

		err := service.ValidateSignatories(
			[]*iam.User{client, admin},
			prj,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects signatory who is not a project member", func(t *testing.T) {
		prj := newTestProject(t)

		client := newTestUser(
			t,
			iam.UserID("client-1"),
			iam.RoleClient,
		)

		staff := newTestUser(
			t,
			iam.UserID("not-a-member"),
			iam.RoleStaff,
		)

		err := service.ValidateSignatories(
			[]*iam.User{client, staff},
			prj,
		)

		if err != ErrContractSignatoryNotProjectMember {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractSignatoryNotProjectMember,
				err,
			)
		}
	})

	t.Run("requires a client signatory", func(t *testing.T) {
		prj := newTestProject(t)

		staff := newTestUser(
			t,
			iam.UserID("staff-1"),
			iam.RoleStaff,
		)

		admin := newTestUser(
			t,
			iam.UserID("admin-1"),
			iam.RoleAdministrator,
		)

		err := service.ValidateSignatories(
			[]*iam.User{staff, admin},
			prj,
		)

		if err != ErrContractMustHaveClientSignatory {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractMustHaveClientSignatory,
				err,
			)
		}
	})

	t.Run("requires a team signatory", func(t *testing.T) {
		prj := newTestProject(t)

		client := newTestUser(
			t,
			iam.UserID("client-1"),
			iam.RoleClient,
		)

		err := service.ValidateSignatories(
			[]*iam.User{client},
			prj,
		)

		if err != ErrContractMustHaveTeamSignatory {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractMustHaveTeamSignatory,
				err,
			)
		}
	})

	t.Run("accepts multiple clients and team members", func(t *testing.T) {
		prj := newTestProject(t)

		client1 := newTestUser(
			t,
			iam.UserID("client-1"),
			iam.RoleClient,
		)

		client2 := newTestUser(
			t,
			iam.UserID("admin-1"),
			iam.RoleClient,
		)

		staff := newTestUser(
			t,
			iam.UserID("staff-1"),
			iam.RoleStaff,
		)

		err := service.ValidateSignatories(
			[]*iam.User{client1, client2, staff},
			prj,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
