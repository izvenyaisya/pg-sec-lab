package policy

type Metadata struct {
	System  string `yaml:"system"`
	Version string `yaml:"version"`
}

type TenantConfig struct {
	Enabled bool   `yaml:"enabled"`
	Setting string `yaml:"setting"`
}

type RolePrivilege struct {
	Object  string   `yaml:"object"`
	Actions []string `yaml:"actions"`
}

type Role struct {
	Login       bool            `yaml:"login"`
	CanCreateDB bool            `yaml:"can_create_db"`
	Members     []string        `yaml:"members"`
	Privileges  []RolePrivilege `yaml:"privileges"`
}

type MaskRule struct {
	Column     string `yaml:"column"`
	Expression string `yaml:"expression"`
	ExposedAs  string `yaml:"exposed_as"`
}

type RLSConfig struct {
	Enabled      bool   `yaml:"enabled"`
	SelectPolicy string `yaml:"select_policy"`
}

type TablePolicy struct {
	RLS   RLSConfig  `yaml:"rls"`
	Masks []MaskRule `yaml:"masks"`
}

type Policy struct {
	Metadata Metadata               `yaml:"metadata"`
	Tenants  TenantConfig           `yaml:"tenants"`
	Roles    map[string]Role        `yaml:"roles"`
	Tables   map[string]TablePolicy `yaml:"tables"`
}
