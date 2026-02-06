package service

import (
	"fmt"

	"gitee.com/fatzeng/switch-admin/internal/repository"
)

// NamespaceMembersService 命名空间跟用户的关联业务
type NamespaceMembersService struct {
	namespaceMembersRepo *repository.NamespaceMembersRepository
}

func NewNamespaceMembersService() *NamespaceMembersService {
	return &NamespaceMembersService{
		namespaceMembersRepo: repository.NewNamespaceMembersRepository(),
	}
}

func (n *NamespaceMembersService) FindApprovePermissionsByNamespaceTag(namespaceTag string) ([]uint, error) {
	userIds := make([]uint, 0)
	owners, err := n.namespaceMembersRepo.FindApprovePermissionsByNamespaceTag(namespaceTag)
	if err != nil {
		return userIds, fmt.Errorf("FindApprovePermissionsByNamespaceTag has error: %w search param is: %v", err, namespaceTag)
	}
	if len(owners) == 0 {
		return userIds, fmt.Errorf("FindApprovePermissionsByNamespaceTag result is empty search param is: %v", namespaceTag)
	}

	var approverUsers []uint
	var approverUsersDistinct = make(map[uint]struct{})
	for _, owner := range owners {
		if _, ok := approverUsersDistinct[owner.UserId]; ok {
			continue
		} else {
			approverUsers = append(approverUsers, owner.UserId)
			approverUsersDistinct[owner.UserId] = struct{}{}
		}
	}
	if len(approverUsers) <= 0 {
		return userIds, fmt.Errorf("FindApprovePermissionsByNamespaceTag approverUsers result is empty search param is: %v", namespaceTag)
	}

	return approverUsers, nil
}
