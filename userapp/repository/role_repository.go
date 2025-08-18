package repository

// Fake repository (later می‌تونی DB بزنی)
type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
    return &RoleRepository{}
}

func (r *RoleRepository) GetRoleByUserID(userID string) string {
    // برای MVP یه role ثابت برمی‌گردونیم
    return "contributor"
}

