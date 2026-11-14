package helper

import "sort"

// PermissionSet menyimpan pemetaan role ke himpunan permission miliknya di memori (O(1) lookup).
type PermissionSet struct {
	byRole map[string]map[string]struct{}
}

// NewPermissionSet mengubah hasil query RoleRepository menjadi nested map untuk pencarian instan.
func NewPermissionSet(raw map[string][]string) *PermissionSet {
	byRole := make(map[string]map[string]struct{}, len(raw))
	for role, permissions := range raw {
		set := make(map[string]struct{}, len(permissions))
		for _, permission := range permissions {
			set[permission] = struct{}{}
		}
		byRole[role] = set
	}
	return &PermissionSet{byRole: byRole}
}

// Can memeriksa apakah role memiliki permission tertentu.
// Menerapkan prinsip FAIL CLOSED: jika receiver nil, role tidak dikenal, atau permission tidak ada,
// fungsi SELALU mengembalikan false.
func (p *PermissionSet) Can(role, permission string) bool {
	if p == nil {
		return false
	}
	permissions, ok := p.byRole[role]
	if !ok {
		return false
	}
	_, granted := permissions[permission]
	return granted
}

// PermissionsOf mengembalikan daftar permission milik suatu role secara terurut.
func (p *PermissionSet) PermissionsOf(role string) []string {
	result := []string{}
	if p == nil {
		return result
	}
	for permission := range p.byRole[role] {
		result = append(result, permission)
	}
	sort.Strings(result)
	return result
}

// KnownRoles mengembalikan seluruh nama role yang terdaftar di sistem secara terurut.
func (p *PermissionSet) KnownRoles() []string {
	result := []string{}
	if p == nil {
		return result
	}
	for role := range p.byRole {
		result = append(result, role)
	}
	sort.Strings(result)
	return result
}

// IsKnownRole memastikan apakah sebuah nama role terdaftar di dalam sistem.
func (p *PermissionSet) IsKnownRole(role string) bool {
	if p == nil {
		return false
	}
	_, ok := p.byRole[role]
	return ok
}
