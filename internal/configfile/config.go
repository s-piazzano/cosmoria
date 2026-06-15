package configfile

type Config struct {
	Project     string           `yaml:"project"`
	Tenants     []TenantConfig   `yaml:"tenants"`
	Collections []CollectionConfig `yaml:"collections"`
	Roles       []RoleConfig     `yaml:"roles"`
}

type TenantConfig struct {
	Name string `yaml:"name"`
}

type CollectionConfig struct {
	Name   string       `yaml:"name"`
	Schema SchemaConfig `yaml:"schema"`
}

type SchemaConfig struct {
	Fields []FieldConfig `yaml:"fields"`
}

type FieldConfig struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
}

type RoleConfig struct {
	Name        string             `yaml:"name"`
	Permissions []PermissionConfig `yaml:"permissions"`
}

type PermissionConfig struct {
	Resource string `yaml:"resource"`
	Action   string `yaml:"action"`
}
