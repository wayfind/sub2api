//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type groupMemberUserRepoStub struct {
	UserRepository
	user      *User
	addErr    error
	removeErr error
	added     [][2]int64
	removed   [][2]int64
}

func (s *groupMemberUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	return s.user, nil
}

func (s *groupMemberUserRepoStub) AddGroupToAllowedGroups(_ context.Context, userID, groupID int64) error {
	if s.addErr != nil {
		return s.addErr
	}
	s.added = append(s.added, [2]int64{userID, groupID})
	return nil
}

func (s *groupMemberUserRepoStub) RemoveGroupFromUserAllowedGroups(_ context.Context, userID, groupID int64) error {
	if s.removeErr != nil {
		return s.removeErr
	}
	s.removed = append(s.removed, [2]int64{userID, groupID})
	return nil
}

type groupMemberGroupRepoStub struct {
	GroupRepository
	group *Group
}

func (s *groupMemberGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	return s.group, nil
}

type allowedGroupsUpdateUserRepoStub struct {
	UserRepository
	user *User
}

func (s *allowedGroupsUpdateUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	clone := *s.user
	clone.AllowedGroups = append([]int64(nil), s.user.AllowedGroups...)
	return &clone, nil
}

func (s *allowedGroupsUpdateUserRepoStub) Update(_ context.Context, user *User) error {
	clone := *user
	clone.AllowedGroups = append([]int64(nil), user.AllowedGroups...)
	s.user = &clone
	return nil
}

func TestAdminService_AddGroupMember_InvalidatesUserAuthCache(t *testing.T) {
	userRepo := &groupMemberUserRepoStub{user: &User{ID: 42}}
	groupRepo := &groupMemberGroupRepoStub{group: &Group{ID: 55, Visibility: VisibilityPrivate}}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             userRepo,
		groupRepo:            groupRepo,
		authCacheInvalidator: invalidator,
	}

	err := svc.AddGroupMember(context.Background(), 55, 42)
	require.NoError(t, err)
	require.Equal(t, [][2]int64{{42, 55}}, userRepo.added)
	require.Equal(t, []int64{42}, invalidator.userIDs)
}

func TestAdminService_RemoveGroupMember_InvalidatesUserAuthCache(t *testing.T) {
	userRepo := &groupMemberUserRepoStub{}
	groupRepo := &groupMemberGroupRepoStub{group: &Group{ID: 55, Visibility: VisibilityPrivate}}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             userRepo,
		groupRepo:            groupRepo,
		authCacheInvalidator: invalidator,
	}

	err := svc.RemoveGroupMember(context.Background(), 55, 42)
	require.NoError(t, err)
	require.Equal(t, [][2]int64{{42, 55}}, userRepo.removed)
	require.Equal(t, []int64{42}, invalidator.userIDs)
}

func TestAdminService_GroupMemberWriteFailure_DoesNotInvalidateAuthCache(t *testing.T) {
	writeErr := errors.New("db write failed")
	groupRepo := &groupMemberGroupRepoStub{group: &Group{ID: 55, Visibility: VisibilityPrivate}}

	t.Run("add", func(t *testing.T) {
		invalidator := &authCacheInvalidatorStub{}
		svc := &adminServiceImpl{
			userRepo:             &groupMemberUserRepoStub{user: &User{ID: 42}, addErr: writeErr},
			groupRepo:            groupRepo,
			authCacheInvalidator: invalidator,
		}

		err := svc.AddGroupMember(context.Background(), 55, 42)
		require.ErrorIs(t, err, writeErr)
		require.Empty(t, invalidator.userIDs)
	})

	t.Run("remove", func(t *testing.T) {
		invalidator := &authCacheInvalidatorStub{}
		svc := &adminServiceImpl{
			userRepo:             &groupMemberUserRepoStub{removeErr: writeErr},
			groupRepo:            groupRepo,
			authCacheInvalidator: invalidator,
		}

		err := svc.RemoveGroupMember(context.Background(), 55, 42)
		require.ErrorIs(t, err, writeErr)
		require.Empty(t, invalidator.userIDs)
	})
}

func TestAdminService_UpdateUserAllowedGroups_InvalidatesUserAuthCache(t *testing.T) {
	userRepo := &allowedGroupsUpdateUserRepoStub{user: &User{
		ID:            42,
		Status:        StatusActive,
		AllowedGroups: []int64{9},
	}}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             userRepo,
		authCacheInvalidator: invalidator,
	}
	updatedGroups := []int64{9, 55}

	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{AllowedGroups: &updatedGroups})
	require.NoError(t, err)
	require.Equal(t, []int64{9, 55}, userRepo.user.AllowedGroups)
	require.Equal(t, []int64{42}, invalidator.userIDs)
}
