package repository

// Fake repository 
type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
    return &RoleRepository{}
}

func (r *RoleRepository) GetRoleByUserID(userID string) string {
    // 
    return "contributor"
}

